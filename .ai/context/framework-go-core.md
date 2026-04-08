Framework: go-core

go-core acts as the service runtime framework.

Responsibilities:

- application bootstrap
- configuration loading
- logger
- database initialization
- lifecycle management
- grpc server
- http gateway
- observability

Example usage:

Database access:

core.SQLByName("transaction")

Logger:

core.Logger()

Transaction:

dbtx.WithTx(ctx, db, func(txCtx context.Context) error {...})

Rules:

services must not create database connections.

services must not create custom loggers.

services must not implement their own bootstrap logic.