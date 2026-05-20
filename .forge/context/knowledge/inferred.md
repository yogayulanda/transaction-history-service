---
id: knowledge.inferred
title: Inferred Knowledge Ledger
type: knowledge
status: confirmed
confidence: high
source: human
owner: TBD
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
| I-001 | Service uses clean architecture (handler → service → repository) | `internal/{handler,service,repository}/` + `.ai/architecture.md` | TBD | 2026-05-20 | inferred |
| I-002 | Framework-first runtime via go-core | `go.mod` replace directive + `.ai/decisions.md` | TBD | 2026-05-20 | inferred |
| I-003 | gRPC-first contract; HTTP is gateway projection | `proto/history/v1/history.proto` + `gen/go/` + `.ai/decisions.md` | TBD | 2026-05-20 | inferred |
| I-004 | Transactional ingestion across 3 tables | `internal/repository/transaction_sql.go` + `.ai/decisions.md` #3 | TBD | 2026-05-20 | inferred |
| I-005 | Error contract via `coreerrors.AppError` + `coreerrors.ToGRPC` | `internal/service/errors.go` + `.ai/integrations.md` | TBD | 2026-05-20 | inferred |
| I-006 | SQL Server is sole persistence backend | `gorm.io/driver/sqlserver` in go.mod + `.ai/integrations.md` | TBD | 2026-05-20 | inferred |
| I-007 | DB tables: `transaction_histories`, `transaction_history_details`, `transaction_history_status_events` | `migrations/transaction/0001_init_transaction_history.up.sql` + `.ai/architecture.md` | TBD | 2026-05-20 | inferred |
| I-008 | Producers: `trxFinance`, `ms-liquiditas`, `agent-payment-purchase` | `README.md` Sumber Data section | TBD | 2026-05-20 | inferred |
| I-009 | Tests are co-located using sqlmock for DB layer | `internal/**/*_test.go` files + `go-sqlmock` in go.mod | TBD | 2026-05-20 | inferred |
| I-010 | Build automation via Makefile (`make proto`, `make test`, `make run`) | `Makefile` | TBD | 2026-05-20 | inferred |
