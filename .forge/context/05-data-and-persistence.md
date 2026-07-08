---
id: system.data-persistence
title: Data and Persistence
type: system
system_type: service
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: migrations/transaction/0001_init_transaction_history.up.sql }
  - { type: code, ref: migrations/transaction/0003_error_definitions.up.sql }
  - { type: code, ref: internal/repository/transaction_sql.go }
  - { type: code, ref: .env.example }
owner: unresolved
updated: 2026-07-06
---

# Data and Persistence

## Primary Store
- SQL Server database named `transaction_history` configured through `DB_LIST=transaction_history` and `DB_TRANSACTION_HISTORY_*` environment variables.

## Main Tables
- `dbo.transaction_histories`: core transaction row keyed by string `id`.
- `dbo.transaction_history_details`: one-to-one metadata table keyed by `transaction_id`.
- `dbo.transaction_history_status_events`: append-only status-event table keyed by identity `id`.
- `dbo.transaction_error_definitions`: service-owned error presentation lookup table.

## Important Constraints
- Unique index on `transaction_histories.reference_id`.
- Check constraints enforce allowed `product_group`, `transaction_route`, `direction`, and `status_code` values.
- Amount, fee, and total_amount must be non-negative.
- `transaction_history_details` and `transaction_history_status_events` cascade-delete with the parent history row.
- Core business fields such as `reference_id`, `source_service`, `channel`, `product_group`, `product_type`, `transaction_route`, `status_code`, and `transaction_time` are modeled as first-class columns; `metadata_json` is stored separately in `transaction_history_details` for extensible product-specific metadata.

## Write Behavior
- Repository `Create` inserts history, detail, and initial status event in one DB transaction.
- If `transaction_time` is omitted, repository persists current UTC time.
- If `metadata_json` is omitted, repository persists `{}`.

## Read Behavior
- Detail lookup loads the history row first, then detail metadata; missing detail metadata is normalized to `{}`.
- User-history listing filters by user and optional date/product/route/status fields, sorts by `transaction_time DESC, id DESC`, and fetches `page_size + 1` rows to derive `has_more`.

## Seed and Migration Notes
- Production migration `0002` is intentionally a no-op.
- Local/manual seed data lives under `migrations/dev/dev_seed_transaction_history.sql`.
