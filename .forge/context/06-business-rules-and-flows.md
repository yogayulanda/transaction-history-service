---
id: knowledge.business-rules-flows
title: Business Rules and Flows
type: knowledge
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: internal/service/transaction_service.go }
  - { type: code, ref: internal/service/idempotency.go }
  - { type: code, ref: internal/handler/grpc/handler.go }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
owner: unresolved
updated: 2026-07-06
---

# Business Rules and Flows

## Create Flow
- Required fields: `user_id`, `reference_id`, `channel`, `product_group`, `product_type`, `status_code`, `source_service`, and `currency`.
- `currency` is normalized to uppercase and must be exactly three letters.
- `status_code` is normalized to uppercase.
- `amount`, `fee`, and `total_amount` must be non-negative.
- `metadata_json` must be a valid JSON object; empty input becomes `{}`.

## Read Flow
- `GetUserHistory` rejects invalid RFC3339 dates, invalid cursor values, and reversed date ranges.
- Default page size is `20`; max page size is `100`.
- `next_cursor` is computed as `offset + page_size` when there are more rows.

## Idempotent Kafka Create Flow
- Kafka inbound uses `CreateTransactionHistoryIdempotent`.
- A duplicate `reference_id` is a no-op only when the incoming payload is equivalent to the already persisted record, including normalized JSON metadata and transaction time.
- A duplicate `reference_id` with different business content becomes `ErrConflictingDuplicateReference`.

## Source and Producer Policy
- Code-level producer allowlist currently accepts only `trxFinance`, `ms-liquiditas`, and `ms-agent-payment-purchase`.
- `CreateTransactionHistory` remains the documented fallback/manual ingestion path even after Kafka inbound support exists.
