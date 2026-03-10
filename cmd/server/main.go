package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	coreapp "github.com/yogayulanda/go-core/app"
	coreconfig "github.com/yogayulanda/go-core/config"
	coremigration "github.com/yogayulanda/go-core/migration"
	coreserver "github.com/yogayulanda/go-core/server"
	coregateway "github.com/yogayulanda/go-core/server/gateway"
	coregrpc "github.com/yogayulanda/go-core/server/grpc"

	historyv1 "github.com/yogayulanda/transaction-history-service/gen/go/history/v1"
	serviceapp "github.com/yogayulanda/transaction-history-service/internal/app"

	_ "github.com/go-sql-driver/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yogayulanda/go-core/logger"
	"google.golang.org/grpc"
)

func main() {
	// Context untuk graceful shutdown (Ctrl+C / SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load config (akan baca .env jika ada, tidak override env production)
	cfg, err := coreconfig.Load(
		coreconfig.WithDotEnv(".env"),
	)
	if err != nil {
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	if err := coremigration.AutoRunUp(cfg); err != nil {
		panic(err)
	}

	// Init core platform (logger, SQL, Redis, Kafka, lifecycle)
	core, err := coreapp.New(cfg)
	if err != nil {
		panic(err)
	}

	// Init service container (repository, usecase, handler)
	svcApp, err := serviceapp.New(core)
	if err != nil {
		panic(err)
	}

	// Setup gRPC server
	grpcServer, err := coregrpc.New(core)
	if err != nil {
		panic(err)
	}

	grpcServer.Register(func(s *grpc.Server) {
		historyv1.RegisterHistoryServiceServer(
			s,
			svcApp.HistoryHandler,
		)
	})

	// Setup HTTP Gateway
	gatewayServer, err := coregateway.New(
		core,
		func(ctx context.Context, mux *runtime.ServeMux) error {
			return historyv1.RegisterHistoryServiceHandlerServer(
				ctx,
				mux,
				svcApp.HistoryHandler,
			)
		},
	)
	if err != nil {
		panic(err)
	}

	httpEndpoints, grpcMethods := coreserver.DescribeFromProto(
		historyv1.File_proto_history_v1_history_proto,
		true,
	)

	core.Logger().Info(context.Background(), "available endpoints",
		logger.Field{Key: "http_endpoints", Value: httpEndpoints},
		logger.Field{Key: "grpc_methods", Value: grpcMethods},
		logger.Field{Key: "http_port", Value: cfg.HTTP.Port},
		logger.Field{Key: "grpc_port", Value: cfg.GRPC.Port},
	)

	// Start servers
	go coreserver.LogStartupReadiness(
		ctx,
		core.Logger(),
		cfg.GRPC.Port,
		cfg.HTTP.Port,
		10*time.Second,
	)

	if err := coreserver.Run(ctx, core, grpcServer, gatewayServer); err != nil {
		panic(err)
	}
}
