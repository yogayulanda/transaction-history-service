package main

import (
	"context"
	"fmt"
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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yogayulanda/go-core/logger"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := coreconfig.Load(
		coreconfig.WithDotEnv(".env"),
	)
	if err != nil {
		exitWithError("load config", err)
	}

	if err := cfg.Validate(); err != nil {
		exitWithError("validate config", err)
	}

	if err := coremigration.AutoRunUp(cfg); err != nil {
		exitWithError("run migration", err)
	}

	core, err := coreapp.New(ctx, cfg)
	if err != nil {
		exitWithError("initialize core app", err)
	}

	svcApp, err := serviceapp.New(core)
	if err != nil {
		core.Logger().Error(ctx, "service bootstrap failed", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	grpcServer, err := coregrpc.New(core)
	if err != nil {
		core.Logger().Error(ctx, "grpc server bootstrap failed", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	grpcServer.Register(func(s *grpc.Server) {
		historyv1.RegisterHistoryServiceServer(
			s,
			svcApp.HistoryHandler,
		)
	})

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
		core.Logger().Error(ctx, "http gateway bootstrap failed", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
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

	go coreserver.LogStartupReadiness(
		ctx,
		core.Logger(),
		cfg.GRPC.Port,
		cfg.HTTP.Port,
		10*time.Second,
		true,
	)

	if err := coreserver.Run(ctx, core, grpcServer, gatewayServer); err != nil {
		core.Logger().Error(ctx, "server stopped with error", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}
}

func exitWithError(step string, err error) {
	fmt.Fprintf(os.Stderr, "transaction-history-service: %s: %v\n", step, err)
	os.Exit(1)
}
