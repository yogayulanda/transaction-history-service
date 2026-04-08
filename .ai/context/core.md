SYSTEM CONTEXT:

go-core is the application kernel (framework + infrastructure).

Architecture:
- handler → service → repository
- service must NOT access DB directly
- repository handles persistence

Framework Usage:
- DB: core.SQLByName("<name>")
- Logger: core.Logger()
- Lifecycle: core.Lifecycle()
- Transaction: dbtx.WithTx

Bootstrap Flow:
config → validate → migration → core app → service wiring → grpc → gateway → run

Notes:
- transaction-history-service is built on top of go-core
- infrastructure is owned by go-core, not by service