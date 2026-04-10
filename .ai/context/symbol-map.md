Symbol Map

Purpose:
Provide a quick reference of important functions and where they are implemented.

This allows AI to locate code faster without scanning entire files.

Service Bootstrap

cmd/server/main.go

Main function:
main()

Responsibilities:
start server
initialize go-core runtime


Application Wiring

internal/app/app.go

Main functions:

New()


Handler Layer

internal/handler/grpc/handler.go

Handlers:

CreateTransactionHistory()
GetUserHistory()
GetTransactionHistoryDetail()

Tests:

internal/handler/grpc/handler_test.go


Service Layer

internal/service/transaction_service.go

Main functions:

CreateTransactionHistory()
GetTransactionHistoryDetail()
GetUserHistory()

Error constructors:

internal/service/errors.go

NewInvalidInputError()
NewDuplicateReferenceIDError()
NewNotFoundError()
NewInternalError()

Tests:

internal/service/transaction_service_test.go


Repository Layer

internal/repository/transaction_sql.go

Main functions:

Create()
FindDetailByID()
ListByUser()

Tests:

internal/repository/transaction_sql_test.go


Domain

internal/domain/transaction.go

Types:

TransactionHistory
TransactionHistoryDetail
CreateTransactionHistoryInput
ListUserHistoryFilter

internal/domain/repository.go

Interfaces:

TransactionRepository

internal/domain/errors.go

ErrTransactionNotFound
ErrInvalidStatus


Database

migrations/

Responsibilities:
database schema management


gRPC API

proto/

Definitions for:

transaction service
history request
history response