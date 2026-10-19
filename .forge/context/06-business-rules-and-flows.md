---
id: knowledge.business-rules-flows
title: Business Rules and Flows
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: internal/service/status_event_service.go }
  - { type: code, ref: internal/handler/grpc/history_handler.go }
  - { type: code, ref: internal/handler/kafka/status_event.go }
  - { type: code, ref: internal/backfill/backfill.go }
owner: unresolved
updated: 2026-10-16
---

# Business Rules and Flows

## Status Event Flow
- Validate restored CloudEvents ID/type/time/subject/sequence and snake_case snapshot payload.
- Validate product/category/provider/app/source, direction, status, currency, decimal format, lengths, and product-specific detail.
- Normalize and mask source/destination values before persistence; destination HMAC input includes number type.
- Duplicate event IDs are acknowledged. Lower/equal versions are stale. Higher versions replace the snapshot; final-to-non-final regression is applied and logged.
- Validation failures are sent to `<topic>.dlq`; DB errors remain retryable.

## Customer Read Flow
- Resolve owner from claims only (`Subject` and `app_name` attribute).
- A transaction owned by someone else is indistinguishable from an unknown ID (`CORE-1004`).
- Dates use Jakarta calendar days, default latest 30 days, max 92 days, inclusive end converted to next-day exclusive.
- Cursor is base64-encoded `(transaction_at, id)`; page size defaults to 20, max 50.

## Backoffice Search
- Require the configured role; require a supported criterion.
- Normalize and hash destination number before repository search.
- Emit `LogEvent` with audit category/action/outcome, criterion names, hash prefix only, and result count.

## Backfill
- Source records reuse status-event validation and PII preparation.
- UUIDv5 event identity is derived from source name and source ID; event version is zero and source system is `backfill`.
- Existing transaction IDs are untouched; invalid rows are counted by reason.
- Report aggregates by Jakarta day and product type using exact decimal arithmetic.
