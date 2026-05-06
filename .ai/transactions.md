# Transactions and Persistence

## Transaction Boundary

- `repository.Create` uses `dbtx.WithTx(ctx, sqlDB, fn)`.
- Create-side writes must succeed/fail together.

## SQL Guarantees (Migration)

- Status values constrained at DB level.
- Unique index on `reference_id` (`UX_transaction_histories_reference_id`).
- Indexed read paths:
  - `user_id, transaction_time`
  - `status_code, transaction_time`
  - `product_type, transaction_time`

## Error Mapping

- SQL duplicate key -> service duplicate-reference AppError.
- Missing transaction -> `ErrTransactionNotFound`.

## Non-Goals

- No network calls inside DB transaction.
- No business-policy branching inside raw SQL strings.
