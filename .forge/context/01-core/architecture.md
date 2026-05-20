---
id: core.architecture
title: Architecture Context
type: core
status: inferred
confidence: high
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: doc, ref: .ai/architecture.md }
  - { type: doc, ref: .ai/context.md }
  - { type: doc, ref: .ai/decisions.md }
  - { type: code, ref: migrations/transaction/0001_init_transaction_history.up.sql }
  - { type: code, ref: migrations/transaction/0003_error_definitions.up.sql }
owner: unresolved
updated: 2026-05-20
---

# Architecture

## Style

Clean architecture with strict 3-layer separation: `handler → service → repository`. Framework-first runtime: bootstrap, transport middleware, auth, and infra are delegated to `go-core`.

## Major Components

| Component | Responsibility |
|---|---|
| `cmd/server` | Bootstrap entrypoint, transport startup |
| `internal/app` | DI wiring (SQL, cache, Kafka publisher, handler construction) |
| `internal/handler/grpc` | Transport validation, mapping, error conversion via `coreerrors.ToGRPC` |
| `internal/service` | Business validation, orchestration, error shaping |
| `internal/repository` | SQL persistence, query, transaction boundary |
| `internal/domain` | Entities, repository interface, sentinels |
| `proto/history/v1` | Proto contract (source of truth for API) |
| `gen/go/history/v1` | Generated gRPC + grpc-gateway code |
| `migrations/transaction` | DB migrations (goose) |

## Runtime Bootstrap Flow

1. `cmd/server/main.go`
2. `config.Load` + `cfg.Validate`
3. `migration.AutoRunUp`
4. `coreapp.New(ctx, cfg)`
5. `internal/app.New(core)` wires dependencies
6. Register gRPC + HTTP gateway handlers
7. `server.Run(...)` with graceful shutdown

## High-Level Data Flow

```
Producer (see core.product → producers)
   │
   ▼ gRPC / HTTP gateway
handler/grpc ─► service ─► repository ─► SQL Server
                                          │
                                          ▼
                              [create-flow transaction]
                              transaction_histories
                              transaction_history_details
                              transaction_history_status_events
                              ─────────────────────────────────
                              [migration-seeded, read-only runtime]
                              transaction_error_definitions
```

## Data Ownership

| Table | Runtime Role | Migration |
|---|---|---|
| `dbo.transaction_histories` | `operational-write` · `transactional-write` | `0001_init_transaction_history.up.sql` |
| `dbo.transaction_history_details` | `operational-write` · `transactional-write` | `0001_init_transaction_history.up.sql` |
| `dbo.transaction_history_status_events` | `operational-write` · `transactional-write` | `0001_init_transaction_history.up.sql` |
| `dbo.transaction_error_definitions` | `migration-seeded` · `lookup/reference` | `0003_error_definitions.up.sql` |

Create-flow transaction boundary: first 3 tables only. `transaction_error_definitions` is seeded by migration and read at runtime for error resolution — not written by runtime code.

## External Integrations

| Integration | Type | Status |
|---|---|---|
| `go-core` framework | Internal Go module (replace `../go-core`) | Required |
| SQL Server | Database (`transaction_history`) | Required |
| Redis | Cache | Optional |
| Kafka | Publisher | Optional (when enabled) |

## Architectural Decisions

| ADR | Status | Topic |
|---|---|---|
| `ADR-0001` | accepted | forge-context-engine adoption |

> Additional architectural decisions discovered during init (framework-first runtime, SQL Server as source of truth, transactional ingestion, gRPC-first contract, app-error contract via go-core, fallback ingestion) are sourced from `.ai/decisions.md` but have NOT been formalized as ADR files. They are tracked as candidate ADRs in `knowledge/assumptions.md` (priority `important`) and will become ADRs only when written.
