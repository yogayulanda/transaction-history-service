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
Producer (trxFinance, ms-liquiditas, agent-payment-purchase)
   │
   ▼ gRPC / HTTP gateway
handler/grpc ─► service ─► repository ─► SQL Server
                                          │
                                          ▼
                                    transaction_histories
                                    transaction_history_details
                                    transaction_history_status_events
```

## Data Ownership

- `dbo.transaction_histories` (main)
- `dbo.transaction_history_details`
- `dbo.transaction_history_status_events`

## External Integrations

| Integration | Type | Status |
|---|---|---|
| `go-core` framework | Internal Go module (replace `../go-core`) | Required |
| SQL Server | Database (`transaction_history`) | Required |
| Redis | Cache | Optional |
| Kafka | Publisher | Optional (when enabled) |

## Major Architectural Decisions

Tracked in `knowledge/decisions/`:

- ADR-0001: forge-context-engine adoption
- ADR-0002 (planned): framework-first runtime via go-core (sourced from `.ai/decisions.md`)
- ADR-0003 (planned): SQL Server as source of truth
- ADR-0004 (planned): transactional ingestion (3-table single transaction)
- ADR-0005 (planned): gRPC-first contract; HTTP is gateway projection
- ADR-0006 (planned): app error contract integration via go-core
- ADR-0007 (planned): `CreateTransactionHistory` retained as fallback/manual ingestion path
