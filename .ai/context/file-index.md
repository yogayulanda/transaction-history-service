Project File Index

Purpose:
Provide quick reference for where code is located.

Main service layers:

Handler layer

internal/handler/grpc/handler.go
internal/handler/grpc/handler_test.go

Responsible for:

- gRPC handlers
- request validation
- error mapping via errors.ToGRPC

Service layer

internal/service/transaction_service.go
internal/service/transaction_service_test.go
internal/service/errors.go

Responsible for:

- business logic
- orchestration
- error contract construction

Repository layer

internal/repository/transaction_sql.go
internal/repository/transaction_sql_test.go

Responsible for:

- database queries
- persistence
- dbtx.WithTx transaction orchestration

Models

internal/domain/transaction.go
internal/domain/repository.go
internal/domain/errors.go

Responsible for:

- domain entities
- repository interface contract
- domain error sentinels

API definitions

proto/

Responsible for:

- gRPC definitions
- API contracts

Server bootstrap

cmd/server/main.go

Responsible for:

- service startup
- go-core integration

Application wiring

internal/app/app.go

Responsible for:

- dependency injection
- component wiring

Database migrations

migrations/