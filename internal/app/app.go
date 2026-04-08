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

	// 1️⃣ Ambil SQL Server dari core
	sqlDB := core.SQLByName("transaction")
	if sqlDB == nil {
		return nil, fmt.Errorf("transaction database not initialized")
	}

	// 1b️⃣ Pastikan MySQL juga tersedia (akan dipakai di phase berikutnya)
	mysqlDB := core.SQLByName("history")
	if mysqlDB == nil {
		return nil, fmt.Errorf("history database not initialized")
	}
	_ = mysqlDB

	// 2️⃣ Wrap sql.DB → GORM
	gormDB, err := gorm.Open(
		sqlserver.New(sqlserver.Config{
			Conn: sqlDB.DB,
		}),
		&gorm.Config{
			PrepareStmt: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("init gorm failed: %w", err)
	}

	// 3️⃣ Ambil abstraction dari core
	var cacheClient cache.Cache
	if core.RedisCache() != nil {
		cacheClient = core.RedisCache()
	}

	var publisher messaging.Publisher

	var log logger.Logger = core.Logger()

	// 4️⃣ Init service layer
	txService := service.NewTransactionService(
		repository.NewTransactionRepository(gormDB, sqlDB.DB),
		cacheClient,
		publisher,
		log,
	)

	// 5️⃣ Init gRPC handler
	historyHandler := handlergrpc.NewHistoryHandler(txService)

	return &App{
		HistoryHandler: historyHandler,
	}, nil
}
