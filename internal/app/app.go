package app

import (
	"context"
	"fmt"
	"time"

	coreapp "github.com/yogayulanda/go-core/app"
	"github.com/yogayulanda/go-core/cache"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"

	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"

	"github.com/yogayulanda/transaction-history-service/internal/domain"
	handlergrpc "github.com/yogayulanda/transaction-history-service/internal/handler/grpc"
	handlerkafka "github.com/yogayulanda/transaction-history-service/internal/handler/kafka"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/service"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type App struct {
	HistoryHandler historyv1.HistoryServiceServer
	KafkaInbound   *kafkaInboundRunner
}

func New(core *coreapp.App) (*App, error) {
	// Primary store for transaction history.
	sqlDB := core.SQLByName("transaction_history")
	if sqlDB == nil {
		return nil, fmt.Errorf("transaction_history database not initialized")
	}

	gormDB, err := gorm.Open(
		sqlserver.New(sqlserver.Config{
			Conn: sqlDB.DB,
		}),
		&gorm.Config{
			PrepareStmt:    true,
			TranslateError: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("init gorm failed: %w", err)
	}

	var cacheClient cache.Cache
	if core.RedisCache() != nil {
		cacheClient = core.RedisCache()
	}

	var publisher messaging.Publisher
	if core.Config().Kafka.Enabled {
		pub, err := core.NewKafkaPublisher()
		if err != nil {
			return nil, fmt.Errorf("init kafka publisher failed: %w", err)
		}
		publisher = pub
	}

	var log logger.Logger = core.Logger()

	txRepo := repository.NewTransactionRepository(gormDB, sqlDB.DB, log)

	var errorResolver service.ErrorDefinitionResolver
	if errDefRepo, ok := txRepo.(domain.ErrorDefinitionRepository); ok {
		resolver := service.NewDBErrorDefinitionResolver(errDefRepo, log)
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

		refreshCtx, cancelRefresh := context.WithCancel(context.Background())
		resolver.StartAutoRefresh(refreshCtx, 60*time.Second)
		core.Lifecycle().Register(func(ctx context.Context) error {
			cancelRefresh()
			return nil
		})

		errorResolver = resolver
	}

	txService := service.NewTransactionService(
		txRepo,
		cacheClient,
		publisher,
		log,
		errorResolver,
	)

	historyHandler := handlergrpc.NewHistoryHandler(txService)
	kafkaInbound, err := buildKafkaInbound(core, txService, publisher, log)
	if err != nil {
		return nil, err
	}

	return &App{
		HistoryHandler: historyHandler,
		KafkaInbound:   kafkaInbound,
	}, nil
}

func buildKafkaInbound(
	core *coreapp.App,
	txService *service.TransactionService,
	publisher messaging.Publisher,
	log logger.Logger,
) (*kafkaInboundRunner, error) {
	inboundCfg := loadKafkaInboundConfig()
	if !inboundCfg.Enabled {
		return nil, nil
	}
	if publisher == nil {
		return nil, fmt.Errorf("kafka inbound enabled but kafka publisher is not initialized")
	}

	handler := handlerkafka.NewTransactionCreatedHandler(txService, publisher, log)
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
