---
id: knowledge.assumptions
title: Assumptions Ledger
type: knowledge
status: confirmed
confidence: high
source: human
owner: unresolved
updated: 2026-05-20
---

# Assumptions

Temporary assumptions ledger. Not a basis for final decisions.

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Ledger — append-only |
| AI writable | Yes — AI adds new assumptions here, **never** to `source: human` files |
| Human confirmation | Not required to add; **required** to promote to `inferred`/`confirmed` |
| Populated | Throughout project lifecycle, especially during planning & init |

## Rules

- Each assumption has owner & status.
- Validated → promote to `inferred.md` (with evidence) → `confirmed` (with entry in `confirmations.md`).
- Invalidated → move to `unknowns.md` or mark `deprecated`.

## Entries

| ID | Assumption | Priority | Owner | Created | Status | Notes |
|---|---|---|---|---|---|---|
| A-001 | Existing `.ai/` folder content is accurate as of 2026-05-20 (used as input for inference) | important | unresolved | 2026-05-20 | assumption | Re-confirm with code review |
| A-002 | Tier `standard` is the right fit; this single-service repo doesn't need `advanced` | informational | unresolved | 2026-05-20 | assumption | Re-evaluate if observability/security layers needed |
| A-003 | `default_mode: implementation` matches the dominant work pattern | informational | unresolved | 2026-05-20 | assumption | Adjust if planning/review work dominates |

## Candidate ADRs (Not Yet Written)

These were referenced by `.ai/decisions.md` and confirmed inferentially against code, but no formal ADR file exists. Promote to a real ADR when authoring.

| ID | Topic | Priority | Owner | Created | Status | Notes |
|---|---|---|---|---|---|---|
| A-ADR-002 | Framework-first runtime via go-core | important | unresolved | 2026-05-20 | assumption | Source: `.ai/decisions.md` #1; evidence in `go.mod` replace + `internal/app/app.go` |
| A-ADR-003 | SQL Server as source of truth for transaction history | important | unresolved | 2026-05-20 | assumption | Source: `.ai/decisions.md` #2; evidence in `go.mod` `gorm.io/driver/sqlserver` + migrations |
| A-ADR-004 | Transactional ingestion across 3 operational tables in single DB transaction (`transaction_histories`, `transaction_history_details`, `transaction_history_status_events`) | important | unresolved | 2026-05-20 | assumption | Source: `.ai/decisions.md` #3; evidence in `internal/repository/transaction_sql.go`; `transaction_error_definitions` is migration-seeded, not part of this transaction |
| A-ADR-005 | gRPC-first contract; HTTP is grpc-gateway projection | important | unresolved | 2026-05-20 | assumption | Source: `.ai/decisions.md` #4; evidence in `proto/` + `gen/` |
| A-ADR-006 | App-error contract integration via `coreerrors.AppError` + `coreerrors.ToGRPC` | important | unresolved | 2026-05-20 | assumption | Source: `.ai/decisions.md` #5 |
| A-ADR-007 | `CreateTransactionHistory` retained as fallback/manual ingestion path | important | unresolved | 2026-05-20 | assumption | Source: `.ai/decisions.md` #6 |
