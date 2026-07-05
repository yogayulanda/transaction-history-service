---
id: system.interfaces-contracts
title: Interfaces and Contracts
type: system
system_type: service
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: code, ref: internal/handler/grpc/handler.go }
  - { type: doc, ref: docs/api.md }
owner: unresolved
updated: 2026-07-06
---

# Interfaces and Contracts

## Public Service Contract
- `HistoryService` exposes three RPCs:
- `GetUserHistory`
- `GetTransactionHistoryDetail`
- `CreateTransactionHistory`

## HTTP Gateway Mapping
- `GET /v1/history`
- `GET /v1/history/{id}`
- `POST /v1/transaction-histories`
- Repository docs also treat `GET /health`, `GET /ready`, `GET /version`, and `GET /metrics` as available runtime endpoints via `go-core`.

## Request and Response Rules
- `GetUserHistory` requires `user_id`.
- `start_date` and `end_date` must be RFC3339 if provided, and `start_date <= end_date`.
- `page_size` defaults to `20` and is clamped to `100`.
- `cursor` is currently a stringified non-negative integer offset.
- `status_code` is the canonical gRPC field for status filtering; HTTP exposes it as `statusCode`.
- `CreateTransactionHistory` accepts business fields plus `metadata_json`; missing `transaction_time` is allowed.

## Status Enum
- Supported lifecycle values are `CREATED`, `PENDING`, `PROCESSING`, `SUCCESS`, `FAILED`, `REVERSED`, and `EXPIRED`.
- `TRANSACTION_STATUS_CODE_UNSPECIFIED` means no status filter on list requests.

## Event Contract
- Optional Kafka inbound expects `event_type=transaction.created`, `event_version=1`, producer allowlist membership, and a payload that maps onto `CreateTransactionHistoryInput`.
- Invalid/malformed/non-allowed events are treated as non-retryable and sent to DLQ when possible.
