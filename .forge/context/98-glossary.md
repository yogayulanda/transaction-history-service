---
id: knowledge.glossary
title: Glossary
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: doc, ref: docs/api.md }
  - { type: doc, ref: docs/events/transaction-status-v1.md }
  - { type: code, ref: proto/history/v1/history.proto }
owner: unresolved
updated: 2026-10-16
---

# Glossary

- `transaction-history-service`: Store for transaction snapshots and ordered status history.
- `transaction_id`: Snapshot primary key and Kafka partition key.
- `reference_no`: Optional external transaction reference used for support search.
- `source_system`: Config-allowlisted event producer identity.
- `detail`: Filtered, versioned product-specific JSON fields.
- `CloudEvents`: Event metadata carried in `ce_*` Kafka headers and restored to handler context.
- `status_version`: Per-transaction monotonic sequence used to choose the current snapshot.
- `destination_no_hash`: Type-prefixed HMAC-SHA256 of the normalized destination value.
- `DLQ`: `<HISTORY_STATUS_TOPIC>.dlq`; invalid records carry `x-dlq-reason`.
- `CORE-1000`: Invalid request response code with optional field details.
- `CORE-1001`: Unauthenticated response code.
- `CORE-1003`: Forbidden response code.
- `CORE-1004`: Owner-scoped not-found response code.
- `CORE-9000`: Internal service failure response code.
- `STATUS-*`: Database-backed status label keys with code fallbacks.
- `MessageProvider`: go-core interface for user-facing message overrides.
