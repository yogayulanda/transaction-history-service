# AI Session Context

## Repository

- Service: `transaction-history-service`
- Language: Go `1.24.x`
- Framework runtime: `go-core` (local replace `../go-core`)

## Current Architecture

- Layering: `handler -> service -> repository`
- Transport: gRPC + grpc-gateway HTTP projection
- Persistence: SQL Server (`transaction_history`)

## Runtime and Infra Ownership

- Service owns: business validation, domain mapping, persistence queries.
- go-core owns: bootstrap, config validation, auth middleware, signature middleware, pprof toggle, logging pipeline, metrics/tracing, lifecycle.

## Current Behavior Snapshot

- `CreateTransactionHistory` is internal fallback/manual ingestion API.
- `GetUserHistory` uses offset-based cursor placeholder (numeric string).
- Date range validation (`start_date <= end_date`) is enforced in handler.
- Create flow writes history + detail + initial status event in one SQL transaction.
- gRPC errors are mapped using `coreerrors.ToGRPC`.

## Canonical AI Context

- Canonical docs are root `.ai/*.md` files listed in `.ai/config.yaml`.
- No legacy `.ai/context/*` dependency.
