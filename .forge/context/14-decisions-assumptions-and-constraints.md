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
updated: 2026-07-06
---

# Decisions Assumptions and Constraints

## Confirmed Decisions
- `CreateTransactionHistory` is intentionally retained as a fallback/manual ingestion path.
- `reference_id` uniqueness is enforced at the database layer and treated as a business idempotency boundary.
- Kafka inbound duplicate events become no-ops only for equivalent persisted content; conflicting duplicates are rejected.
- Error presentation can be overridden from DB-backed active definitions without changing code.

## Confirmed Constraints
- Pagination is offset-based today even though the API uses `cursor` naming.
- Kafka inbound producer trust is currently a hardcoded name allowlist.
- `metadata_json` must remain a JSON object, not an arbitrary JSON value.
- The service requires the `transaction_history` SQL database to be initialized at bootstrap.

## Assumptions Still in Use
- `go-core` remains the owner of transport-level auth, readiness, metrics, and shared infrastructure behavior.
- Deployment topology and ownership will be provided from outside this repository when needed.
