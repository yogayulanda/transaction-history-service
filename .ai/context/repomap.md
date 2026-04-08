Repository Map

Purpose:
Provide a quick structural overview of the service so AI can locate relevant code without scanning the entire repository.

Service: transaction-history-service

Framework:
go-core

Entry Points

cmd/

Main service startup.
Bootstraps the application using go-core.

Typical file:

cmd/server/main.go


Application Wiring

internal/app/

Responsible for:

service initialization
dependency injection
component wiring


Handler Layer

internal/handler/

Handles transport layer.

Responsibilities:

HTTP handlers
gRPC handlers
request validation

Calls service layer.


Service Layer

internal/service/

Contains business logic.

Responsibilities:

orchestrate operations
apply business rules
call repositories


Repository Layer

internal/repository/

Handles database persistence.

Responsibilities:

SQL queries
data storage
data retrieval


Domain Models

internal/model/

Defines entities and domain models.


API Definitions

proto/

Defines gRPC contracts.

Generated code located in:

gen/


Database Migrations

migrations/

Contains database schema migrations.


External Dependencies

go-core framework

Provides:

bootstrap
database
logger
grpc
http gateway
observability


Typical Feature Flow

Example: fetch transaction history

handler
↓
service
↓
repository
↓
database


Important Rules

handlers must not access repository directly

service must not depend on transport

repository must not contain business logic