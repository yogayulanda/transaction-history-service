# Architecture

## Layer Model

- `handler`:
  - validates and parses request input
  - maps transport models <-> domain models
  - converts errors to gRPC via `coreerrors.ToGRPC`
- `service`:
  - business validation and orchestration
  - maps repository errors to service error contract
- `repository`:
  - SQL persistence/query logic
  - transaction boundary and DB logging

## Data Ownership

- Main table: `dbo.transaction_histories`
- Detail table: `dbo.transaction_history_details`
- Status event table: `dbo.transaction_history_status_events`

## Design Rules

- No transport logic in service/repository.
- No SQL from handler/service.
- No cross-layer shortcuts.
- Prefer additive changes; avoid broad refactors.
