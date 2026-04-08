Feature Map

Purpose:
Provide a high level map of which features are implemented in which modules.

This allows AI to quickly locate the correct implementation area without scanning the entire repository.

Core Features

Transaction History

location:

internal/repository/transaction_repository.go
internal/service/transaction_service.go
internal/handler/transaction_handler.go

responsibility:

store transaction records
retrieve transaction history by user_id


Service Bootstrap

location:

cmd/server/main.go
internal/app/

responsibility:

initialize service
configure dependencies
bootstrap go-core runtime


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