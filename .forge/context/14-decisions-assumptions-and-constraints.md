---
id: knowledge.decisions-assumptions-constraints
title: Decisions Assumptions and Constraints
type: knowledge
status: confirmed
confidence: medium
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: code, ref: internal/service/idempotency.go }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
  - { type: code, ref: internal/repository/transaction_sql.go }
owner: unresolved
updated: 2026-10-15
---

# Decisions Assumptions and Constraints

## Confirmed Decisions
- D3: Adopt the go-core v1.2.1 HTTP envelope without a compatibility shim or `gateway.Raw`.
- D4: Register `TRH-VAL-002`, `TRH-DB-001`, and `TRH-REC-001`; active `transaction_error_definitions.user_message` values provide message overrides. `details_json` is retained but unused.
- D8: The service is not deployed; environment validation and live SQL Server/HTTP smoke checks are required before first deployment.
- D9: Service field validation uses `CORE-1000`; `TRH-VAL-001` is retired and its existing database row remains unused. Migration 0004 adds the Indonesian CORE-1000 message.
- D10: A registered `AppError` is Kafka non-retryable only when its definition is `ClassFailed` and not retryable.
- D11: This service stores transaction history and does not emit go-core transaction logs. Create identity is propagated through the observability field bag; `gtid` is propagated from gRPC/Kafka and is not set by the service.
- `CreateTransactionHistory` is intentionally retained as a fallback/manual ingestion path.
- `reference_id` uniqueness is enforced at the database layer and treated as a business idempotency boundary.
- Kafka inbound duplicate events become no-ops only for equivalent persisted content; conflicting duplicates are rejected.
- Error message presentation can be overridden from DB-backed active definitions without changing code.

## Confirmed Constraints
- Pagination is offset-based today even though the API uses `cursor` naming.
- Kafka inbound producer trust is currently a hardcoded name allowlist.
- `metadata_json` must remain a JSON object, not an arbitrary JSON value.
- The service requires the `transaction_history` SQL database to be initialized at bootstrap.

## Assumptions Still in Use
- `go-core` remains the owner of transport-level auth, readiness, metrics, and shared infrastructure behavior.
- Deployment topology and ownership will be provided from outside this repository when needed.
