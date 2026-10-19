---
id: system.interfaces-contracts
title: Interfaces and Contracts
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: code, ref: proto/history/v1/backoffice.proto }
  - { type: doc, ref: docs/api.md }
  - { type: doc, ref: docs/events/transaction-status-v1.md }
owner: unresolved
updated: 2026-10-16
---

# Interfaces and Contracts

## Customer API
- `HistoryService`: `ListTransactions`, `GetTransaction`.
- HTTP: `GET /v1/transactions`, `GET /v1/transactions/{transaction_id}`.
- No customer request accepts `user_id` or `app_name`; owner derives from claims.
- Dates are `YYYY-MM-DD` in Asia/Jakarta. Default range is 30 days; maximum is 92 days; end date is inclusive.
- Page size defaults to 20 and is capped at 50. Cursor is opaque base64 over `(transaction_at, id)`.
- Money values are decimal strings. Detail uses product-specific oneof messages.

## Backoffice API
- `HistoryBackofficeService.SearchTransactions` is gRPC-only.
- Requires role `HISTORY_BACKOFFICE_ROLE` (default `backoffice`).
- At least one of transaction ID, reference number, destination number/type, or app/user is required.
- Destination searches normalize and HMAC before lookup; maximum page size is 100.

## Event Contract
- CloudEvents attributes are restored by go-core and read from context; no payload fallback is used.
- `ce_id` is a UUID, `ce_subject` equals `transaction_id`, and positive `ce_sequence` orders snapshots.
- Full fields, product schemas, and retry/DLQ behavior are documented in `docs/events/transaction-status-v1.md`.

## Errors and Messages
- `CORE-1000`: invalid request; `CORE-1001`: unauthenticated; `CORE-1003`: forbidden; `CORE-1004`: not found; `CORE-9000`: internal.
- HTTP uses the shared response envelope. Active DB messages override defaults; status labels use `STATUS-<STATUS>` keys.
