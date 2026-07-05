---
id: system.runtime-deployment-config
title: Runtime Deployment and Config
type: system
system_type: service
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: .env.example }
  - { type: code, ref: Makefile }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-07-06
---

# Runtime Deployment and Config

## Startup Sequence
- Load `.env`
- Validate config
- Auto-run migrations when enabled
- Build `go-core` app
- Build service app wiring
- Start gRPC server and HTTP gateway
- Start Kafka inbound only when enabled and available

## Core Config Surface
- Service identity: `SERVICE_NAME=transaction-history-service`
- Ports: `GRPC_PORT`, `HTTP_PORT`
- Shutdown: `SHUTDOWN_TIMEOUT`
- DB migration: `MIGRATION_AUTO_RUN`, `MIGRATION_DB`, `MIGRATION_DIR`
- SQL Server connection: `DB_TRANSACTION_HISTORY_*`
- Observability: `OTEL_EXPORTER_OTLP_ENDPOINT`, `TRACE_SAMPLING_RATIO`, `HTTP_PPROF_ENABLED`
- Auth/signature: `INTERNAL_JWT_*`, `AUTH_SIGNATURE_*`
- Kafka: `KAFKA_ENABLED`, `KAFKA_BROKERS`, `KAFKA_CLIENT_ID`, `KAFKA_USERNAME`, `KAFKA_PASSWORD`, `KAFKA_INBOUND_ENABLED`

## Developer Commands
- `make proto` / `make generate`
- `make test`
- `make run`

## Deployment Unknowns
- This repo does not include deployment manifests, CI/CD pipelines, or environment-specific topology docs.
