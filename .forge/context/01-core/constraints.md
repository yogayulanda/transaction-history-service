---
id: core.constraints
title: Hard Constraints
type: core
status: inferred
confidence: high
source: ai
evidence:
  - { type: doc, ref: .ai/security.md }
  - { type: doc, ref: .ai/integrations.md }
  - { type: doc, ref: README.md }
  - { type: code, ref: go.mod }
  - { type: code, ref: migrations/transaction/0001_init_transaction_history.up.sql }
  - { type: code, ref: migrations/transaction/0003_error_definitions.up.sql }
  - { type: code, ref: internal/service/transaction_service.go }
owner: unresolved
updated: 2026-05-20
---

# Constraints

"Must" boundaries — non-negotiable limits. Highest authority after live code & ADR.

## Required Technology

| Constraint | Source |
|---|---|
| Go 1.24.3+ | `go.mod` |
| `go-core` framework (local replace `../go-core`) | `go.mod` replace directive |
| SQL Server (DB: `transaction_history`) | `.ai/integrations.md`, README |
| gRPC + grpc-gateway | proto + gen layout |

## Auth & Security (Framework-Controlled)

- Internal JWT verification: env-toggled (`INTERNAL_JWT_ENABLED`).
- Signature middleware: env-toggled (`AUTH_SIGNATURE_ENABLED`).
- HTTP pprof endpoint: env-toggled (`HTTP_PPROF_ENABLED`).
- These are security-impacting; never weaken in handler code.

## Service Rules

- Never log secrets or raw auth credentials.
- Keep error responses sanitized via app error contract.
- Do not weaken auth checks in handlers.
- Do not bypass framework middleware with custom transport handling.

## Data Integrity

- `reference_id` must be unique across all producers (enforced by `UX_transaction_histories_reference_id`).
- API status enum values must stay consistent with SQL constraints.
- **Create-flow transaction writes to 3 operational tables:** `transaction_histories`, `transaction_history_details`, `transaction_history_status_events` — inside one DB transaction via `dbtx.WithTx`. No partial writes allowed.
- `transaction_error_definitions` is **migration-seeded and read at runtime** for error definition resolution. It is **not** part of runtime create-flow transactional writes.
- Transaction detail rows are 1:1 with `transaction_histories` and cascade-delete via FK.
- `transaction_history_status_events` is append-only.

## Implicit Enum Constraints (DB-Enforced)

Enforced by SQL `CHECK` constraints in `transaction_histories` (and mirrored in `transaction_history_status_events` for status):

| Field | Allowed Values | DB Constraint |
|---|---|---|
| `product_group` | `ppob`, `transfer`, `cash` | `CK_transaction_histories_product_group` |
| `transaction_route` | `internal`, `bifast`, `rtol`, `switching`, `partner_h2h` | `CK_transaction_histories_transaction_route` |
| `direction` | `debit`, `credit` | `CK_transaction_histories_direction` |
| `status_code` | `CREATED`, `PENDING`, `PROCESSING`, `SUCCESS`, `FAILED`, `REVERSED`, `EXPIRED` | `CK_transaction_histories_status_code` |
| `to_status_code` (status events) | Same 7 values as `status_code` | `CK_transaction_history_status_events_to_status_code` |
| `from_status_code` (status events) | Same 7 values OR `NULL` | `CK_transaction_history_status_events_from_status_code` |

Any code path that produces a value outside these sets will be rejected by the database.

## Implicit Numeric Constraints

- `amount`, `fee`, `total_amount` — non-negative (DB constraint `CK_transaction_histories_amount_non_negative`).
- All monetary fields stored as `BIGINT` (smallest currency unit).
- `currency` is `CHAR(3)` — ISO-4217 code (e.g. `IDR`).

## Required Field Validation (Service Layer)

Enforced by empty-check in `internal/service/transaction_service.go` → `sanitizeCreateInput`:

**Required (empty-check returns error):**
`user_id`, `reference_id`, `channel`, `product_group`, `product_type`, `status_code`, `source_service`, `currency`

**Trimmed but NOT required (no empty-check):**
`direction`, `transaction_route`, `external_ref_id`, `error_code`, `error_message`

**Numeric constraints (service-layer):**
`amount >= 0`, `fee >= 0`, `total_amount >= 0`

**`transaction_time`:** not validated at service layer. Repository falls back to `time.Now().UTC()` when zero-value (see `internal/repository/transaction_sql.go` lines 102–105).

**`metadata_json`:** defaults to `"{}"` if empty; must be a valid JSON object if provided.

**`currency`:** must be a 3-letter alphabetic code (validated by `isAlphaString`).

> **Context accuracy note:** Three validation layers exist independently:
> - *Service layer* — explicit empty-checks and format checks in `sanitizeCreateInput`. Only the fields above are enforced here.
> - *DB constraints* — SQL `CHECK` constraints enforce enum values for `product_group`, `transaction_route`, `direction`, `status_code` (see Implicit Enum Constraints above). A field can pass service validation and still be rejected by the DB if its value is outside the allowed set.
> - *Repository fallback* — `transaction_time` is set to current UTC time by the repository when the caller passes a zero-value. This is not a service-layer default; it happens at persistence time.

## Error Code Catalog (DB-Seeded)

Pre-defined user-facing errors in `dbo.transaction_error_definitions` (seeded by `0003_error_definitions.up.sql`):

| Error Code | Indicates |
|---|---|
| `TRH-VAL-001` | Generic validation failure |
| `TRH-VAL-002` | Duplicate `reference_id` |
| `TRH-DB-001` | Transaction history not found |
| `TRH-REC-001` | System busy / retryable |

Service code MUST use these codes when raising `coreerrors.AppError` for the corresponding cases.

## Field Schema Rule

Core business fields stay as columns (not in `metadata_json`):
- `reference_id`, `external_ref_id`, `source_service`, `channel`, `product_group`, `product_type`, `transaction_route`, `direction`, `status_code`, `transaction_time`, `amount`, `fee`, `total_amount`, `currency`

`metadata_json`:
- Must be a valid JSON object (default `'{}'`).
- Only for product-specific attributes that don't fit core columns.

## Build & Tooling

- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway` required for codegen.
- Generated proto code must be regenerated when `proto/` changes.
- `go.mod` and `go.sum` must remain synchronized (`go mod tidy`).
