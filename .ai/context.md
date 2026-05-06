# transaction-history-service · AI Context

## Purpose

Service to ingest and query transaction history records.

- Transport: gRPC + HTTP gateway
- Storage: SQL Server (`transaction_history`)
- Framework runtime: `go-core`

## Core Runtime Flow

1. `cmd/server/main.go`
2. `config.Load` + `cfg.Validate`
3. `migration.AutoRunUp`
4. `coreapp.New(ctx, cfg)`
5. `internal/app.New(core)` wires repository/service/handler
6. register gRPC + HTTP gateway handlers
7. `server.Run(...)` with graceful shutdown

## Must-Know Constraints

- Layering is strict: `handler -> service -> repository`.
- SQL writes for create flow are transactional via `dbtx.WithTx`.
- API status enum values must stay consistent with SQL constraints.
- `GetUserHistory` cursor is currently numeric offset placeholder.
- This service owns domain behavior; go-core owns infrastructure behavior.

## Primary Source Files

- `cmd/server/main.go`
- `internal/app/app.go`
- `internal/handler/grpc/handler.go`
- `internal/service/transaction_service.go`
- `internal/repository/transaction_sql.go`
- `proto/history/v1/history.proto`
- `migrations/transaction/*.sql`
