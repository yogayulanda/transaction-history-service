# Modules

## Entry and Wiring

- `cmd/server/main.go`: bootstrap + transport startup
- `internal/app/app.go`: DI wiring (`SQLByName`, cache, Kafka publisher, handler construction)

## API Layer

- `proto/history/v1/history.proto`
- `internal/handler/grpc/handler.go`

RPCs:
- `CreateTransactionHistory`
- `GetUserHistory`
- `GetTransactionHistoryDetail`

## Domain + Service

- `internal/domain/*.go`: entities, repository interface, sentinels
- `internal/service/transaction_service.go`: orchestration entrypoints
- `internal/service/errors.go`: app-level error contract helpers

## Persistence

- `internal/repository/transaction_sql.go`
  - `Create` (transactional insert into 3 tables)
  - `FindDetailByID`
  - `ListByUser` (offset pagination)

## Tests

- `internal/handler/grpc/handler_test.go`
- `internal/service/transaction_service_test.go`
- `internal/repository/transaction_sql_test.go`
