# Key Decisions

1. Framework-first runtime
- Service delegates bootstrap, transport middleware, auth/signature enforcement, and infra wiring to `go-core`.

2. SQL Server as source of truth
- Transaction history persistence is SQL-backed with explicit schema constraints and indexes.

3. Transactional ingestion
- `CreateTransactionHistory` writes history/detail/status_event in one DB transaction.

4. gRPC-first contract
- HTTP API is gateway projection of proto contract, not an independent contract.

5. App error contract integration
- Service errors are built as `coreerrors.AppError` and converted by handler via `coreerrors.ToGRPC`.
- Error metadata follows go-core taxonomy fields (`domain`, `category`, `number`, `finality`, `retryable`).

6. Fallback ingestion stays explicit
- `CreateTransactionHistory` remains an internal fallback/manual ingestion endpoint while primary ingest may evolve separately.
