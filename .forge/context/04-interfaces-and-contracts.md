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
  - { type: code, ref: go.mod }
owner: unresolved
updated: 2026-10-15
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
- HTTP success responses use the go-core v1.2.1 envelope: `status`, `response_code`, `message`, `request_id`, `timestamp`, and `data`.
- HTTP errors use the same envelope metadata; `details` is emitted for `CORE-1000` field violations. The detail RPC's protobuf `data` remains nested inside the envelope's `data`.
- Documented envelope examples are derived from go-core v1.2.1 gateway code and are pending live smoke verification.

## Error Contract
- `CORE-1000`: invalid request, HTTP 400, with optional field details.
- `TRH-VAL-002`: duplicate reference, HTTP 400.
- `TRH-DB-001`: history not found, HTTP 404.
- `TRH-REC-001` and fallback `CORE-9000`: HTTP 500.
- Active `transaction_error_definitions.user_message` values override registered default messages through the go-core message provider.

## Status Enum
- Supported lifecycle values are `CREATED`, `PENDING`, `PROCESSING`, `SUCCESS`, `FAILED`, `REVERSED`, and `EXPIRED`.
- `TRANSACTION_STATUS_CODE_UNSPECIFIED` means no status filter on list requests.

## Event Contract
- Optional Kafka inbound expects `event_type=transaction.created`, `event_version=1`, producer allowlist membership, and a payload that maps onto `CreateTransactionHistoryInput`.
- Invalid/malformed/non-allowed events are treated as non-retryable and sent to DLQ when possible.
- DLQ header behavior and per-partition ordering are described in `09-errors-and-resilience.md` and `07-integrations-and-dependencies.md`.
