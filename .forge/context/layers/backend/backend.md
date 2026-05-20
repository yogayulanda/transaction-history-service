---
id: layer.backend.content
title: Backend Layer Conventions
type: layer
status: inferred
confidence: high
source: ai
evidence:
  - { type: code, ref: internal/ }
  - { type: code, ref: migrations/ }
  - { type: code, ref: Makefile }
  - { type: code, ref: .env.example }
  - { type: doc, ref: .ai/architecture.md }
  - { type: doc, ref: .ai/conventions.md }
  - { type: doc, ref: .ai/modules.md }
owner: unresolved
updated: 2026-05-20
---

# Backend Layer

Conventions for Go backend code in `internal/`, `cmd/`, `proto/`, `gen/`, plus repo-owned operational concerns (migrations, build, runtime config) that are NOT deployment.

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
migrations/transaction/ DB migrations (goose)
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
3. Service uses pre-defined error codes from `dbo.transaction_error_definitions` (see `core.constraints` → Error Code Catalog).
4. Handler calls `coreerrors.ToGRPC(err)`.
5. Client receives sanitized gRPC status.

## Transactional Writes

`CreateTransactionHistory` writes to **3 operational tables** inside one transaction via `dbtx.WithTx`:
- `transaction_histories`
- `transaction_history_details`
- `transaction_history_status_events`

No partial writes allowed. `transaction_error_definitions` is **not** part of this transaction — it is migration-seeded and read at runtime for error resolution only.

## API Contract

- gRPC contract: `proto/history/v1/history.proto` (source of truth).
- HTTP contract: grpc-gateway annotations in proto. Not a separate spec.
- Health/metrics endpoints provided by go-core: `/health`, `/ready`, `/version`, `/metrics`.

## Configuration

- Env-driven via `go-core` config loader.
- `.env.example` is the contract for required keys.
- Secrets never committed; `.env` is gitignored.

### Config Surface

| Category | Keys |
|---|---|
| Transport | `GRPC_PORT`, `HTTP_PORT` |
| Database | SQL Server connection (per `transaction_history` DB) |
| Migration | `MIGRATION_AUTO_RUN` |
| Auth | `INTERNAL_JWT_*`, `AUTH_SIGNATURE_*` |
| Debug | `HTTP_PPROF_ENABLED` |

## Database Migrations (Repo-Owned)

- Tool: `goose` (`github.com/pressly/goose/v3`)
- Production-bound: `migrations/transaction/`
- Dev seed: `migrations/dev/dev_seed_transaction_history.sql` (local only)
- Naming: `NNNN_description.up.sql` / `NNNN_description.down.sql`
- Auto-run controlled by `MIGRATION_AUTO_RUN` env

### Existing Migrations

- `0001_init_transaction_history.up.sql` — baseline schema (3 tables: histories, details, status_events)
- `0002_seed_transaction_history.up.sql` — production no-op (parity placeholder)
- `0003_error_definitions.up.sql` — `transaction_error_definitions` table + seed rows

## Build Tooling

| Tool | Purpose | How |
|---|---|---|
| `protoc` + plugins | Generate gRPC + gateway code | `make proto` |
| `go build` | Compile binary | implicit in `make run` |
| `go test` | Run tests | `make test` |
| `go mod tidy` | Sync deps | manual |

Required protoc plugins: `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`.

## Generated Code Policy

- `gen/` contains regenerable code from `proto/`.
- Currently committed (per repo convention; ADR pending — see `knowledge/unknowns.md` U-004).
- Regenerate via `make proto` after any `proto/` change.

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
- Hard-coding error user-messages instead of using `transaction_error_definitions` codes.

## Out of Scope (Deployment / Infrastructure)

This repo does NOT own:
- Kubernetes / Helm / Terraform manifests
- CI/CD deployment pipelines
- Cloud resource provisioning
- Secret stores beyond `.env`

Those concerns are tracked in `knowledge/unknowns.md` U-002 (deployment ownership) and likely live in a separate repo.
