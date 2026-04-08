Project File Index

Purpose:
Provide quick reference for where code is located.

Main service layers:

Handler layer

internal/handler/

Responsible for:

- HTTP handlers
- gRPC handlers
- request validation

Service layer

internal/service/

Responsible for:

- business logic
- orchestration

Repository layer

internal/repository/

Responsible for:

- database queries
- persistence

Models

internal/model/

Responsible for:

- domain entities
- database models

API definitions

proto/

Responsible for:

- gRPC definitions
- API contracts

Server bootstrap

cmd/

Responsible for:

- service startup
- go-core integration

Database migrations

migrations/