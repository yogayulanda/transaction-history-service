# Data Flow

## CreateTransactionHistory

1. Handler validates request object + status enum.
2. Service normalizes and validates business-required fields.
3. Repository opens DB transaction.
4. Repository inserts:
   - transaction row
   - detail row
   - initial status event row
5. Repository emits `DBLog`.
6. Service emits `ServiceLog` and `TransactionLog`.
7. Handler returns created `id`.

## GetUserHistory

1. Handler validates user/date/cursor/page size.
2. Handler enforces `start_date <= end_date` when both provided.
3. Service forwards normalized filter to repository.
4. Repository builds filtered SQL query.
5. Query uses `page_size + 1` to compute `has_more`.
6. Handler returns `items`, `has_more`, `next_cursor`.

## GetTransactionHistoryDetail

1. Handler validates `id`.
2. Service calls repository.
3. Repository loads base + detail metadata.
4. Service maps not-found/internal to AppError.
5. Handler converts AppError to gRPC via `coreerrors.ToGRPC`.
