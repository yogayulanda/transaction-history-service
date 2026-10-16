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

	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"

	"github.com/yogayulanda/transaction-history-service/internal/catalog"
	"github.com/yogayulanda/transaction-history-service/internal/domain"
	handlergrpc "github.com/yogayulanda/transaction-history-service/internal/handler/grpc"
	handlerkafka "github.com/yogayulanda/transaction-history-service/internal/handler/kafka"
	"github.com/yogayulanda/transaction-history-service/internal/identity"
	"github.com/yogayulanda/transaction-history-service/internal/pii"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/service"
)

type App struct {
	HistoryHandler      historyv1.HistoryServiceServer
	BackofficeHandler   historyv1.HistoryBackofficeServiceServer
	KafkaInbound        *kafkaInboundRunner
	UsesDefaultResolver bool
}

func New(core *coreapp.App) (*App, error) {
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

	historyRepo := repository.NewHistoryRepository(sqlDB.DB, log)
	var messageResolver *service.DBErrorDefinitionResolver

	if errDefRepo, ok := historyRepo.(domain.ErrorDefinitionRepository); ok {
		resolver := service.NewDBErrorDefinitionResolver(errDefRepo, log)
		messageResolver = resolver
		if err := resolver.Load(context.Background()); err != nil {
			log.LogService(context.Background(), logger.ServiceLog{
				Operation: "error_definition_bootstrap",
				Status:    "failed",
				ErrorCode: "bootstrap_failed",
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		coreerrors.SetMessageProvider(resolver)

		refreshCtx, cancelRefresh := context.WithCancel(context.Background())
		resolver.StartAutoRefresh(refreshCtx, 60*time.Second)
		core.Lifecycle().Register(func(ctx context.Context) error {
			cancelRefresh()
			return nil
		})

	}

	validationCatalog := catalog.DefaultStaticCatalog(os.Getenv("HISTORY_ALLOWED_APPS"), os.Getenv("HISTORY_ALLOWED_SOURCES"))
	identityResolver := identity.NewClaimsResolver(validationCatalog)
	historyService := service.NewHistoryService(historyRepo)
	hashKey, err := loadHistoryHashKey()
	if err != nil {
		return nil, err
	}
	hasher, err := pii.NewHasher(hashKey)
	if err != nil {
		return nil, fmt.Errorf("initialize history hash key: %w", err)
	}
	historyHandler := handlergrpc.NewHistoryHandler(historyService, identityResolver, validationCatalog, messageResolver)
	backofficeHandler := handlergrpc.NewBackofficeHandler(
		historyService,
		validationCatalog,
		messageResolver,
		hasher,
		os.Getenv("HISTORY_BACKOFFICE_ROLE"),
		log,
	)
	statusEventService := service.NewStatusEventService(historyRepo, log)
	kafkaInbound, err := buildKafkaInbound(core, statusEventService, publisher, log)
	if err != nil {
		return nil, err
	}

	return &App{
		HistoryHandler:      historyHandler,
		BackofficeHandler:   backofficeHandler,
		KafkaInbound:        kafkaInbound,
		UsesDefaultResolver: true,
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
	statusEventService *service.StatusEventService,
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

	hasher, err := pii.NewHasher(inboundCfg.HashKey)
	if err != nil {
		return nil, fmt.Errorf("initialize history hash key: %w", err)
	}
	validationCatalog := catalog.NewStaticCatalog(inboundCfg.AllowedApps, inboundCfg.AllowedSources)
	handler := handlerkafka.NewStatusEventHandler(statusEventService, validationCatalog, hasher, publisher, log)
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
