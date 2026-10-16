---
id: system.data-persistence
title: Data and Persistence
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: migrations/transaction/0001_init_transaction_history.up.sql }
  - { type: code, ref: migrations/transaction/0005_core_messages_and_status_labels.up.sql }
  - { type: code, ref: internal/repository/history_sql.go }
owner: unresolved
updated: 2026-10-16
---

# Data and Persistence

## Primary Store
- SQL Server database `transaction_history`.

## Tables
- `dbo.transaction_histories`: snapshot keyed by transaction ID and owned by `(app_name, user_id)`.
- `dbo.transaction_history_details`: optional schema-versioned JSON detail.
- `dbo.transaction_history_status_events`: append-only status events.
- `dbo.transaction_error_definitions`: CORE message and status-label lookup.

## Constraints and Indexes
- Money columns use `DECIMAL(18,2)` and non-negative checks; Go/API money stays string-valued.
- Statuses are `PENDING`, `UNKNOWN`, `SUCCESS`, `FAILED`, `NEED_REVIEW`, `REFUNDED`.
- Event IDs and transaction/status versions are unique.
- Customer list index is `(app_name, user_id, transaction_at DESC, id DESC)` with included summary fields.
- Detail and status-event rows cascade-delete with the transaction.

## Write Behavior
- One SQL transaction covers event-ID deduplication, locked snapshot read, version-based create/update/stale handling, optional detail write, and event append.
- Destination/source numbers are masked or hashed and CICO codes are masked before repository access.
- Transfer notes and PLN prepaid tokens are stored unmasked in schema-filtered detail JSON by design and may be returned only in the owning customer's transaction detail response.
- Account/phone/meter/VA numbers, CICO codes, PLN tokens, and notes must not be logged.
- Backfill inserts only when transaction ID is absent and uses status version zero.

## Read Behavior
- Customer reads always filter by both app and user owner.
- List/search use descending `(transaction_at, id)` keyset paging and fetch one extra record.
- Destination search queries a type-prefixed HMAC hash, not the raw number.
