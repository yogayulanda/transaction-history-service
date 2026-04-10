Feature Map

Purpose:
Provide a high level map of which features are implemented in which modules.

This allows AI to quickly locate the correct implementation area without scanning the entire repository.

Core Features

Transaction History

location:

internal/repository/transaction_sql.go
internal/repository/transaction_sql_test.go
internal/service/transaction_service.go
internal/service/transaction_service_test.go
internal/service/errors.go
internal/handler/grpc/handler.go
internal/handler/grpc/handler_test.go

responsibility:

store transaction records
retrieve transaction history by user_id
retrieve transaction detail by id


Service Bootstrap

location:

cmd/server/main.go
internal/app/app.go

responsibility:

initialize service
configure dependencies
bootstrap go-core runtime


Domain Models

location:

internal/domain/transaction.go
internal/domain/repository.go
internal/domain/errors.go

responsibility:

entities and types
repository interface contract
domain error definitions


Database Persistence

location:

internal/repository/

responsibility:

SQL queries
data persistence


API Layer

location:

proto/
internal/handler/

responsibility:

gRPC contracts
HTTP gateway endpoints


Database Migrations

location:

migrations/

responsibility:

schema evolution


Typical Request Flow

client
↓
handler
↓
service
↓
repository
↓
database