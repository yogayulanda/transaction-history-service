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

internal/app/

Main functions:

New()
initializeDependencies()


Handler Layer

internal/handler/

Typical handlers:

GetTransactionHistory()


Service Layer

internal/service/

Main functions:

GetUserHistory()
CreateTransaction()


Repository Layer

internal/repository/

Main functions:

ListByUser()
InsertTransaction()


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