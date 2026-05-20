---
id: knowledge.inferred
title: Inferred Knowledge Ledger
type: knowledge
status: confirmed
confidence: high
source: human
owner: unresolved
updated: 2026-05-20
---

# Inferred

AI inference ledger quarantined from human-authored facts. **Non-authoritative.**

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Quarantine ledger — append-only |
| AI writable | Yes — new inferences written here, **never** to `source: human` files |
| Human confirmation | Required to promote entry to `confirmed` |
| Populated | During init (brownfield gap-filling) and throughout normal operation |

## Rules

- Each entry **must** have `evidence` (code path, doc, ADR, or external source).
- Entry status ≤ `inferred` until promoted via `confirmations.md`.
- If evidence changes, entry demotes to `assumption` or is removed.

## Entries

| ID | Inference | Evidence | Owner | Created | Status |
|---|---|---|---|---|---|
| I-001 | Service uses clean architecture (handler → service → repository) | `internal/{handler,service,repository}/` + `.ai/architecture.md` | unresolved | 2026-05-20 | inferred |
| I-002 | Framework-first runtime via go-core | `go.mod` replace directive + `.ai/decisions.md` | unresolved | 2026-05-20 | inferred |
| I-003 | gRPC-first contract; HTTP is gateway projection | `proto/history/v1/history.proto` + `gen/go/` + `.ai/decisions.md` | unresolved | 2026-05-20 | inferred |
| I-004 | Transactional ingestion across 3 tables in single create flow | `internal/repository/transaction_sql.go` + `migrations/transaction/0001_init_transaction_history.up.sql` | unresolved | 2026-05-20 | inferred |
| I-005 | Error contract via `coreerrors.AppError` + `coreerrors.ToGRPC` | `internal/service/errors.go` + `.ai/integrations.md` | unresolved | 2026-05-20 | inferred |
| I-006 | SQL Server is sole persistence backend | `gorm.io/driver/sqlserver` in go.mod + `.ai/integrations.md` | unresolved | 2026-05-20 | inferred |
| I-007 | DB has 4 owned tables: `transaction_histories`, `transaction_history_details`, `transaction_history_status_events`, `transaction_error_definitions` | `migrations/transaction/0001_init_transaction_history.up.sql` + `migrations/transaction/0003_error_definitions.up.sql` | unresolved | 2026-05-20 | inferred |
| I-008 | Producers: `trxFinance`, `ms-liquiditas`, `agent-payment-purchase` | `README.md` (Sumber Data section) | unresolved | 2026-05-20 | inferred |
| I-009 | Tests are co-located using `go-sqlmock` for DB layer | `internal/**/*_test.go` files + `go-sqlmock` in go.mod | unresolved | 2026-05-20 | inferred |
| I-010 | Build automation via Makefile (`make proto`, `make test`, `make run`) | `Makefile` | unresolved | 2026-05-20 | inferred |
| I-011 | Implicit enum constraints: `product_group` (3 values), `transaction_route` (5 values), `direction` (2 values), `status_code` (7 values) — enforced at DB level only; `direction` and `transaction_route` are NOT service-required | `migrations/transaction/0001_init_transaction_history.up.sql` CHECK constraints + `internal/service/transaction_service.go` `sanitizeCreateInput` | unresolved | 2026-05-20 | inferred |
| I-012 | Pre-defined error code catalog: `TRH-VAL-001`, `TRH-VAL-002`, `TRH-DB-001`, `TRH-REC-001` | `migrations/transaction/0003_error_definitions.up.sql` seed rows | unresolved | 2026-05-20 | inferred |
| I-013 | `reference_id` is unique-indexed (UX_transaction_histories_reference_id); duplicate triggers `gorm.ErrDuplicatedKey` | `migrations/transaction/0001_init_transaction_history.up.sql` + `internal/service/transaction_service.go` | unresolved | 2026-05-20 | inferred |
| I-014 | `metadata_json` defaults to `'{}'` and stays valid JSON; non-core fields only | `migrations/transaction/0001_init_transaction_history.up.sql` + README | unresolved | 2026-05-20 | inferred |
| I-015 | `transaction_history_details` has FK CASCADE to `transaction_histories.id`; same for `transaction_history_status_events` | `migrations/transaction/0001_init_transaction_history.up.sql` | unresolved | 2026-05-20 | inferred |
| I-016 | Repo does NOT own deployment/IaC; only migrations + build tooling + env config | filesystem scan: no Helm/Terraform/K8s/CI-deploy artifacts | unresolved | 2026-05-20 | inferred |
| I-017 | Table role classification: `transaction_histories`/`details`/`status_events` = `operational-write` + `transactional-write`; `transaction_error_definitions` = `migration-seeded` + `lookup/reference` (not written by runtime create-flow) | `internal/repository/transaction_sql.go` (writes to 3 tables in `dbtx.WithTx`) + `migrations/transaction/0003_error_definitions.up.sql` (INSERT only in migration, no runtime write path) | unresolved | 2026-05-20 | inferred |
