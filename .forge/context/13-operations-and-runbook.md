---
id: knowledge.operations-runbook
title: Operations and Runbook
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/handler/kafka/status_event.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-10-16
---

# Operations and Runbook

## Startup Checks
- If startup exits early, check shared config validation, SQL Server connectivity, migration execution, and the HMAC key configuration.
- When Kafka inbound is enabled, verify `HISTORY_STATUS_TOPIC`, `HISTORY_ALLOWED_SOURCES`, Kafka publisher setup, and valid `SERVICE_NAME`.
- Outside local environment, verify internal JWT is enabled when using the default claims resolver.
- If user-facing messages or status labels look stale, check `transaction_error_definitions` load/refresh logs.

## Message Recovery
- Invalid status events are published to `<HISTORY_STATUS_TOPIC>.dlq` with `x-dlq-reason`.
- Database errors and failed DLQ publishes remain retryable; failed DLQ publication does not acknowledge the input.
- Original DLQ payloads are retained for replay and must never be logged; log only `event_id`, `ce_type`, and a stable reason.
- For the internal fork (I-A), deploy the `.dlq` topic with the source topic's ACL, TLS, and retention controls.
- Inspect stable reason/error codes only; raw identifiers, PLN tokens, and notes are not logged.

## Data and API Recovery
- Validate migrations 0001-0005 against SQL Server, including lock hints, unique conflict behavior, and rollback order.
- Smoke customer list/detail through HTTP and gRPC; verify owner isolation, auth errors, request ID, date bounds, and CORE error envelopes.
- Smoke backoffice search through gRPC only; verify role enforcement and audit records contain no raw destination values.
- Local/manual sample data lives in `migrations/dev/dev_seed_transaction_history.sql`.
- `internal/backfill` is a framework only; this public repository has no source adapter or backfill command.

## Escalation Gaps
- Production ownership, escalation path, and rollback procedure are not documented in this repository.
