---
id: layer.backend.content
title: Backend Layer Conventions
type: layer
status: inferred
confidence: high
source: ai
evidence:
  - { type: code, ref: internal/ }
  - { type: doc, ref: .ai/architecture.md }
  - { type: doc, ref: .ai/conventions.md }
  - { type: doc, ref: .ai/modules.md }
owner: TBD
updated: 2026-05-20
---

# Backend Layer

Conventions for Go backend code in `internal/`, `cmd/`, `proto/`, `gen/`.

## Language & Runtime

- Go 1.24.3
- Module: `github.com/yogayulanda/transaction-history-service`
- Local dependency: `go-core` via replace directive

## Package Layout

```
cmd/server/             entrypoint (main + bootstrap)
internal/app/           DI wiring (single composition root: app.go)
internal/handler/grpc/  transport layer (gRPC + gateway)
internal/service/       business orchestration
internal/repository/    SQL persistence
internal/domain/        entities, repository interface, sentinels
proto/history/v1/       proto source
gen/go/history/v1/      generated gRPC + gateway code
```

## Layer Responsibilities

| Layer | Allowed | Forbidden |
|---|---|---|
| `handler` | Validate input, parse transport models, map errors via `coreerrors.ToGRPC` | SQL queries, business logic |
| `service` | Business validation, orchestration, error shaping, repository calls | Transport types, SQL |
| `repository` | SQL CRUD, transaction management via `dbtx.WithTx` | Business rules, transport types |
| `domain` | Entities, repository interface, error sentinels | Imports from upper layers |

## Error Handling Pattern

1. Domain emits sentinel errors.
2. Service builds `coreerrors.AppError` with metadata: `domain`, `category`, `number`, `finality`, `retryable`.
3. Handler calls `coreerrors.ToGRPC(err)`.
4. Client receives sanitized gRPC status.

## Transactional Writes

`CreateTransactionHistory` writes to 3 tables (`histories`, `details`, `status_events`) inside one transaction via `dbtx.WithTx`. No partial writes allowed.

## API Contract

- gRPC contract: `proto/history/v1/history.proto` (source of truth).
- HTTP contract: grpc-gateway annotations in proto. Not a separate spec.
- Health/metrics endpoints provided by go-core: `/health`, `/ready`, `/version`, `/metrics`.

## Configuration

- Env-driven via `go-core` config loader.
- `.env.example` is the contract for required keys.
- Secrets never committed; `.env` is gitignored.

## Testing Conventions

- Unit tests live next to source: `*_test.go`.
- DB tests use `go-sqlmock`.
- See `layers/testing/testing.md` for strategy.

## Observability

Use go-core log flavors:
- `ServiceLog` — application events
- `DBLog` — database operations
- `TransactionLog` — business transaction lifecycle

Never log raw secrets or auth credentials.

## Anti-Patterns

- Reimplementing auth/signature middleware (framework owns this).
- Direct SQL from handler or service.
- Skipping `coreerrors` for service errors.
- Placing core business fields into `metadata_json`.
- Using HTTP-only response shapes (gRPC contract is canonical).
