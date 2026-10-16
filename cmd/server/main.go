package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	coreapp "github.com/yogayulanda/go-core/app"
	coreconfig "github.com/yogayulanda/go-core/config"
	"github.com/yogayulanda/go-core/logger"
	coremigration "github.com/yogayulanda/go-core/migration"
	coreserver "github.com/yogayulanda/go-core/server"
	serviceapp "github.com/yogayulanda/transaction-history-service/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := coreconfig.Load(coreconfig.WithDotEnv(".env"))
	if err != nil {
		exitWithError("load config", err)
	}
	cfg.Migration.AutoRun = true
	cfg.Migration.DBName = "transaction_history"
	cfg.Migration.Dir = "migrations/transaction"
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
	app, err := serviceapp.New(ctx, core)
	if err != nil {
		core.Logger().Error(ctx, "service bootstrap failed", logger.Field{Key: "error", Value: err.Error()})
		exitWithError("initialize service app", err)
	}
	if err := coreserver.Run(ctx, core, app.Components...); err != nil {
		core.Logger().Error(ctx, "server stopped with error", logger.Field{Key: "error", Value: err.Error()})
		exitWithError("run server", err)
	}
}

func exitWithError(step string, err error) {
	fmt.Fprintf(os.Stderr, "transaction-history-service: %s: %v\n", step, err)
	os.Exit(1)
}
