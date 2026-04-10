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
- error contract and mapper

Example usage:

Database access:

core.SQLByName("transaction_history")

Logger:

core.Logger()

Transaction:

dbtx.WithTx(ctx, db, func(txCtx context.Context) error {...})

Error contract:

coreerrors.New(coreerrors.CodeNotFound, "not found")
coreerrors.Validation("invalid input", coreerrors.Detail{Field: "user_id", Reason: "required"})
coreerrors.ToGRPC(err)

Kafka publisher:

core.NewKafkaPublisher()

Rules:

services must not create database connections.

services must not create custom loggers.

services must not implement their own bootstrap logic.

services must use errors.AppError for service error contract.

services must use errors.ToGRPC for gRPC error mapping.