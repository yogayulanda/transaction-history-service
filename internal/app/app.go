package app

import (
	"fmt"

	coreapp "github.com/yogayulanda/go-core/app"
	"github.com/yogayulanda/go-core/cache"
	"github.com/yogayulanda/go-core/logger"
	"github.com/yogayulanda/go-core/messaging"

	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"

	handlergrpc "github.com/yogayulanda/transaction-history-service/internal/handler/grpc"
	"github.com/yogayulanda/transaction-history-service/internal/repository"
	"github.com/yogayulanda/transaction-history-service/internal/service"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type App struct {
	HistoryHandler historyv1.HistoryServiceServer
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

	txService := service.NewTransactionService(
		repository.NewTransactionRepository(gormDB, sqlDB.DB, log),
		cacheClient,
		publisher,
		log,
	)

	historyHandler := handlergrpc.NewHistoryHandler(txService)

	return &App{
		HistoryHandler: historyHandler,
	}, nil
}
