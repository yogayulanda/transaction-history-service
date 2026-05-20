---
id: system.transaction-history-service
title: "System: transaction-history-service"
type: system
system_type: service
status: inferred
confidence: high
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/handler/grpc/handler.go }
  - { type: code, ref: internal/service/transaction_service.go }
  - { type: code, ref: internal/repository/transaction_sql.go }
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: doc, ref: README.md }
  - { type: doc, ref: .ai/modules.md }
owner: TBD
updated: 2026-05-20
---

# System: transaction-history-service

## Purpose & Responsibilities

Centralized historical store for financial transactions. Owns persistence, querying, and fallback ingestion of transaction records.

Owns:
- Persistence of transaction history records (3-table model).
- Query API for user history list and transaction detail.
- Fallback/manual ingestion path.

Does NOT own:
- Transaction processing logic.
- Aggregate reporting.
- Primary event/Kafka ingestion path.

## Type & Runtime

- `system_type: service`
- Language: Go 1.24.3
- Transports: gRPC (primary), HTTP (grpc-gateway projection)
- Persistence: SQL Server
- Runtime framework: `go-core`

## Public Interfaces

### gRPC RPCs (`proto/history/v1/history.proto`)

| RPC | Status | Notes |
|---|---|---|
| `CreateTransactionHistory` | inferred | Fallback/manual ingestion path |
| `GetUserHistory` | inferred | Cursor is numeric offset placeholder |
| `GetTransactionHistoryDetail` | inferred | Single transaction detail |

### HTTP Gateway (provided by go-core)

| Endpoint | Status | Source |
|---|---|---|
| `GET /health` | inferred | go-core |
| `GET /ready` | inferred | go-core |
| `GET /version` | inferred | go-core |
| `GET /metrics` | inferred | go-core (Prometheus) |

## Dependencies

### Internal (Go modules)

| Dependency | Reference | Type |
|---|---|---|
| `github.com/yogayulanda/go-core` | `replace ../go-core` | Required framework |

### External Services

| Service | Direction | Required? |
|---|---|---|
| SQL Server (`transaction_history` DB) | Outbound (write/read) | Required |
| Redis | Outbound (cache) | Optional |
| Kafka | Outbound (publish) | Optional |

### Producers (Inbound)

Recorded in `01-core/product.md`:
- `trxFinance`
- `ms-liquiditas`
- `agent-payment-purchase`

> These are external systems calling this service. Not Go-module dependencies.

## Layers Touched

References (do NOT copy content):
- `layer.backend.content` — Go conventions, layer rules, error contract
- `layer.infrastructure.content` — migrations, build tooling, env config
- `layer.testing.content` — test strategy

## Implementation-Specific Context

### Database Schema

Owned tables (in `dbo` schema):
- `transaction_histories` — main record
- `transaction_history_details` — per-transaction detail rows
- `transaction_history_status_events` — status change history

### Validation Boundaries

- Date range validation: handler enforces `startDate <= endDate` when both provided.
- Required field validation for create: at service layer (`channel`, `sourceService`, `currency`, `statusCode`, etc.).

### Known Limitations

- Pagination: `GetUserHistory` uses numeric offset as cursor (placeholder).
- Lifecycle: status update flow outside create is not yet complete.
- Aggregate reports: out of scope; downstream consumers handle this.
- Generated proto code: must be regenerated when `proto/` changes.

## Unknowns & Assumptions Specific to This System

See `knowledge/unknowns.md` for unresolved questions specific to this service.
