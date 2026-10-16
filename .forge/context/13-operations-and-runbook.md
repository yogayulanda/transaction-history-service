---
id: knowledge.operations-runbook
title: Operations and Runbook
type: knowledge
status: confirmed
confidence: medium
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-10-15
---

# Operations and Runbook

## Useful Runtime Checks
- If startup exits early, first check config validation, SQL Server connectivity, and migration execution because each aborts the process.
- If Kafka inbound is expected but absent, check both `KAFKA_INBOUND_ENABLED` and the underlying `go-core` Kafka enablement/config because inbound creation requires a publisher.
- If user-facing error text looks stale, check whether `transaction_error_definitions` is populated and whether message-provider load or refresh logs show failures.

## Message Recovery Notes
- Non-retryable Kafka events should land on `transaction-history.transaction.created.dlq`.
- Service-level non-retryable DLQ messages include `x-original-topic` and `x-error-code`; consumer retry-exhaustion DLQ messages include `x-dlq-reason` and `x-original-topic`.
- If DLQ publish fails, the handler returns a retryable error and the original message should be retried by the consumer runtime.
- Before first deployment, verify migrations 0001-0004 and the HTTP error envelope against SQL Server; see `99-open-questions.md`.

## Data Recovery Notes
- Local/manual seed data exists under `migrations/dev/dev_seed_transaction_history.sql`.
- No repository-local reconciliation job or manual backfill tooling was found beyond the fallback create API and optional Kafka ingest.

## Escalation Gaps
- No documented on-call ownership, incident escalation path, or production rollback procedure is present in this repo.
