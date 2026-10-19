---
id: knowledge.errors-resilience
title: Errors and Resilience
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: internal/service/errors.go }
  - { type: code, ref: internal/service/error_definition_resolver.go }
  - { type: code, ref: internal/handler/kafka/status_event.go }
  - { type: code, ref: internal/repository/history_sql.go }
owner: unresolved
updated: 2026-10-16
---

# Errors and Resilience

## API Codes
- `CORE-1000`: invalid request (`InvalidArgument`, HTTP 400).
- `CORE-1001`: unauthenticated (`Unauthenticated`, HTTP 401).
- `CORE-1003`: forbidden (`PermissionDenied`, HTTP 403).
- `CORE-1004`: transaction not found for this owner (`NotFound`, HTTP 404).
- `CORE-9000`: internal failure (`Internal`, HTTP 500).
- Active DB message rows may override default user-facing messages and status labels.

## Kafka Retry and DLQ
- Missing/invalid CloudEvents attributes, malformed payloads, catalog violations, invalid amounts, and detail validation errors are non-retryable.
- Invalid events are published unchanged to `<topic>.dlq` with `x-dlq-reason` so they can be replayed after a consumer fix; the input is acknowledged only after successful publication.
- Repository/database errors are returned for retry.
- A DLQ publish error is returned for retry and prevents input acknowledgment.
- DLQ event bodies are never logged. DLQ-related logs may include only `event_id`, `ce_type`, and a stable reason.
- Deployment requirement for the internal fork (I-A): the `.dlq` topic must have the same ACL, TLS, and retention as its source topic.

## Idempotency and Ordering
- Existing event ID is acknowledged as duplicate.
- A status version no higher than the current version is stale and does not update the header.
- Unique transaction/version conflict with another event ID is logged and acknowledged.
- Higher versions apply even across final-to-non-final regressions; regression is an anomaly signal, not a rejection.
