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
  - { type: code, ref: migrations/transaction/ }
  - { type: doc, ref: README.md }
  - { type: doc, ref: .ai/modules.md }
owner: unresolved
updated: 2026-05-20
---

# System: transaction-history-service

## Purpose & Responsibilities

Centralized historical store for financial transactions. Owns persistence, querying, and fallback ingestion of transaction records.

Owns:
- Persistence of transaction history records (3 operational tables + 1 migration-seeded lookup — see Database Schema).
- Query API for user history list and transaction detail.
- Fallback/manual ingestion path.
- User-facing error message catalog (read-only runtime, seeded by migration).

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

| RPC | Notes |
|---|---|
| `CreateTransactionHistory` | Fallback/manual ingestion path |
| `GetUserHistory` | Cursor is numeric offset placeholder |
| `GetTransactionHistoryDetail` | Single transaction detail |

### HTTP Gateway (provided by go-core)

| Endpoint | Source |
|---|---|
| `GET /health` | go-core |
| `GET /ready` | go-core |
| `GET /version` | go-core |
| `GET /metrics` | go-core (Prometheus) |

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

### Inbound Producers

See `core.product` for the canonical producer list. This system accepts ingestion from those external systems via `CreateTransactionHistory` RPC. Producer list is NOT duplicated here per anti-duplication rule (F7).

## Layers Touched

References (do NOT copy content):
- `layer.backend.content` — Go conventions, layer rules, error contract, migrations, build, runtime config
- `layer.testing.content` — test strategy

> Note: `infrastructure` layer was deactivated during init — repo does not own deployment manifests/IaC. Operational concerns (migrations, build, env config) are part of `layer.backend.content`.

## Implementation-Specific Context

### Database Schema

Owned tables (in `dbo` schema):

| Table | Runtime Role |
|---|---|
| `transaction_histories` | `operational-write` + `transactional-write` — main record, written in create-flow transaction |
| `transaction_history_details` | `operational-write` + `transactional-write` — 1:1 detail, written in same create-flow transaction |
| `transaction_history_status_events` | `operational-write` + `transactional-write` — status event, written in same create-flow transaction; append-only |
| `transaction_error_definitions` | `migration-seeded` + `lookup/reference` — seeded by `0003_error_definitions.up.sql`; read at runtime for error resolution; **not written by runtime create-flow** |

Create-flow transaction boundary: `transaction_histories` + `transaction_history_details` + `transaction_history_status_events` (3 tables). `transaction_error_definitions` is outside this boundary.

Detailed schema constraints (enums, FK cascade, non-negative checks) are documented in `core.constraints`.

### Validation Boundaries

- Date range: handler enforces `startDate <= endDate` when both provided.
- Required field validation for create: service layer (`internal/service/transaction_service.go` → `sanitizeCreateInput`). Required fields list and DB enum constraints documented in `core.constraints`.
- Duplicate `reference_id` mapped to `TRH-VAL-002` via `gorm.ErrDuplicatedKey` detection.

### Service-Specific Operations

> This table is a context aid, not the source of truth. Verify against `internal/service/transaction_service.go` before modifying operation behavior.

| Operation | Status code on success | Status code on logged failure |
|---|---|---|
| `CreateTransactionHistory` | `success` | `failed` (with reason: `validation_failed`, `duplicate_reference`, `repository_error`) |
| `GetTransactionHistoryDetail` | `success` | `failed` (`not_found`, `repository_error`) |
| `GetUserHistory` | `success` | `failed` (`repository_error`) |

These are emitted via `ServiceLog` and `TransactionLog` (go-core flavors).

### Known Limitations

- Pagination: `GetUserHistory` uses numeric offset as cursor (placeholder).
- Lifecycle: status update flow outside create is not yet complete (see `knowledge/unknowns.md` U-009).
- Aggregate reports: out of scope; downstream consumers handle this.
- Generated proto code: must be regenerated when `proto/` changes.

## Unknowns & Assumptions Specific to This System

See `knowledge/unknowns.md` for unresolved questions specific to this service.
