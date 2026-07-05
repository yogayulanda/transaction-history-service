---
id: layer.architecture
title: Architecture
type: layer
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/service/transaction_service.go }
  - { type: code, ref: internal/repository/transaction_sql.go }
owner: unresolved
updated: 2026-07-06
---

# Architecture

## Shape
- Layered Go service with a single process that starts gRPC, grpc-gateway HTTP, optional Kafka inbound consumption, and shared `go-core` runtime components.

## Main Components
- `cmd/server`: loads config, validates config, optionally auto-runs migrations, boots the core app, registers transports, and starts runtime components.
- `internal/app`: assembles SQL Server, GORM repository, optional Redis cache handle, optional Kafka publisher, DB-backed error resolver, gRPC handler, and optional Kafka inbound runner.
- `internal/handler/grpc`: transport mapping and request validation for gRPC and HTTP-gateway traffic.
- `internal/handler/kafka`: event decoding, allowlist checks, idempotent create orchestration, retryability classification, and DLQ publishing.
- `internal/service`: input normalization, business validation, error taxonomy, idempotent duplicate handling, and service/db/transaction logging.
- `internal/repository`: SQL Server persistence through GORM plus explicit transaction handling via `go-core/dbtx`.
- `proto` and `gen/go`: canonical contract source and generated server/gateway bindings.

## Runtime Assembly
- The process always starts gRPC and HTTP gateway servers.
- Kafka inbound starts only when `KAFKA_INBOUND_ENABLED=true` and the core app successfully initializes a Kafka publisher.
- The error-definition resolver loads from DB on startup and refreshes in the background every 60 seconds.

## Design Notes
- Repository writes create three records in one DB transaction: history row, detail row, and initial status event row.
- List pagination is offset-based today even though the API field is named `cursor`.
- Cache and publisher dependencies are optional in wiring; current service logic does not use cache directly.
