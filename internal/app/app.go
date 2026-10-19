package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	coreapp "github.com/yogayulanda/go-core/app"
	coreerrors "github.com/yogayulanda/go-core/errors"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"
	coreserver "github.com/yogayulanda/go-core/server"
	coregateway "github.com/yogayulanda/go-core/server/gateway"
	coregrpc "github.com/yogayulanda/go-core/server/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
	handlergrpc "github.com/yogayulanda/transaction-history-service/internal/handler/grpc"
	handlerkafka "github.com/yogayulanda/transaction-history-service/internal/handler/kafka"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/service"
	"github.com/yogayulanda/transaction-history-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	HistoryHandler    historyv1.HistoryServiceServer
	BackofficeHandler historyv1.HistoryBackofficeServiceServer
	KafkaInbound      *kafkaInboundRunner
	Components        []coreserver.Startable
}

func New(ctx context.Context, core *coreapp.App) (*App, error) {
	// Primary store for transaction history.
	sqlDB := core.SQLByName("transaction_history")
	if sqlDB == nil {
		return nil, fmt.Errorf("transaction_history database not initialized")
	}

	var publisher messaging.Publisher
	if core.Config().Kafka.Enabled {
		pub, err := core.NewKafkaPublisher()
		if err != nil {
			return nil, fmt.Errorf("init kafka publisher failed: %w", err)
		}
		publisher = pub
	}

	log := core.Logger()

	historyRepo := repository.NewTransactionRepository(sqlDB.DB, log)
	messageProvider := service.NewDBMessageProvider(repository.NewMessageRepository(sqlDB.DB), log)
	if err := messageProvider.Load(context.Background()); err != nil {
		log.LogService(context.Background(), logger.ServiceLog{
			Operation: "error_definition_bootstrap",
			Status:    "failed",
			ErrorCode: "bootstrap_failed",
			Metadata: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}
	coreerrors.SetMessageProvider(messageProvider)

	refreshCtx, cancelRefresh := context.WithCancel(context.Background())
	messageProvider.StartAutoRefresh(refreshCtx, 60*time.Second)
	core.Lifecycle().Register(func(ctx context.Context) error {
		cancelRefresh()
		return nil
	})

	validationCatalog := repository.DefaultStaticCatalog(os.Getenv("HISTORY_ALLOWED_APPS"), os.Getenv("HISTORY_ALLOWED_SOURCES"))
	identityResolver := handlergrpc.NewClaimsResolver(validationCatalog)
	historyService := service.NewHistoryService(historyRepo, historyRepo)
	hashKey, err := loadHistoryHashKey()
	if err != nil {
		return nil, err
	}
	hasher, err := utils.NewHasher(hashKey)
	if err != nil {
		return nil, fmt.Errorf("initialize history hash key: %w", err)
	}
	historyHandler := handlergrpc.NewHistoryHandler(historyService, identityResolver, validationCatalog, messageProvider)
	backofficeHandler := handlergrpc.NewBackofficeHandler(
		historyService,
		validationCatalog,
		messageProvider,
		hasher,
		os.Getenv("HISTORY_BACKOFFICE_ROLE"),
		log,
	)
	kafkaInbound, err := buildKafkaInbound(core, historyRepo, publisher, log)
	if err != nil {
		return nil, err
	}
	if core.Config().App.Environment != "local" && !core.Config().Auth.InternalJWT.Enabled {
		return nil, fmt.Errorf("default claims resolver requires internal JWT outside local environment")
	}
	grpcServer, err := coregrpc.New(core)
	if err != nil {
		return nil, err
	}
	grpcServer.Register(func(server *grpc.Server) {
		historyv1.RegisterHistoryServiceServer(server, historyHandler)
		historyv1.RegisterHistoryBackofficeServiceServer(server, backofficeHandler)
	})
	gatewayServer, err := coregateway.New(
		core,
		func(ctx context.Context, mux *runtime.ServeMux) error {
			return historyv1.RegisterHistoryServiceHandlerFromEndpoint(
				ctx,
				mux,
				fmt.Sprintf("localhost:%d", core.Config().GRPC.Port),
				[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			)
		},
	)
	if err != nil {
		return nil, err
	}
	httpEndpoints, grpcMethods := coreserver.DescribeFromProto(historyv1.File_history_v1_history_proto, true)
	log.Info(context.Background(), "available endpoints",
		logger.Field{Key: "http_endpoints", Value: httpEndpoints},
		logger.Field{Key: "grpc_methods", Value: grpcMethods},
		logger.Field{Key: "http_port", Value: core.Config().HTTP.Port},
		logger.Field{Key: "grpc_port", Value: core.Config().GRPC.Port},
	)
	go coreserver.LogStartupReadiness(ctx, log, core.Config().GRPC.Port, core.Config().HTTP.Port, 10*time.Second, core.Config().HTTP.TLSEnabled)
	components := []coreserver.Startable{grpcServer, gatewayServer}
	if kafkaInbound != nil {
		components = append(components, kafkaInbound)
	}

	return &App{
		HistoryHandler:    historyHandler,
		BackofficeHandler: backofficeHandler,
		KafkaInbound:      kafkaInbound,
		Components:        components,
	}, nil
}

func loadHistoryHashKey() ([]byte, error) {
	key, err := hex.DecodeString(os.Getenv("HISTORY_HASH_KEY"))
	if err != nil || len(key) < 32 {
		return nil, fmt.Errorf("HISTORY_HASH_KEY must be hex encoding of at least 32 bytes")
	}
	return key, nil
}

func buildKafkaInbound(
	core *coreapp.App,
	historyRepo domain.TransactionRepository,
	publisher messaging.Publisher,
	log logger.Logger,
) (*kafkaInboundRunner, error) {
	inboundCfg, err := loadKafkaInboundConfig()
	if err != nil {
		return nil, err
	}
	if !inboundCfg.Enabled {
		return nil, nil
	}
	if publisher == nil {
		return nil, fmt.Errorf("kafka inbound enabled but kafka publisher is not initialized")
	}

	hasher, err := utils.NewHasher(inboundCfg.HashKey)
	if err != nil {
		return nil, fmt.Errorf("initialize history hash key: %w", err)
	}
	validationCatalog := repository.NewStaticCatalog(inboundCfg.AllowedApps, inboundCfg.AllowedSources)
	statusEventService := service.NewStatusEventService(historyRepo, validationCatalog, hasher, log)
	handler := handlerkafka.NewStatusEventHandler(statusEventService, publisher, log)
	consumer, err := core.NewKafkaConsumer(
		inboundCfg.Topic,
		inboundCfg.GroupID,
		handler.Handle,
		messaging.WithConsumerConcurrency(inboundCfg.Concurrency),
		messaging.WithConsumerRetry(inboundCfg.RetryMax, inboundCfg.RetryDelay),
		messaging.WithConsumerDLQ(publisher),
	)
	if err != nil {
		return nil, fmt.Errorf("init kafka inbound consumer failed: %w", err)
	}

	return newKafkaInboundRunner(context.Background(), consumer), nil
}
