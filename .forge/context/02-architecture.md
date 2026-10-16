---
id: layer.architecture
title: Architecture
type: layer
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/repository/transaction_repository.go }
  - { type: code, ref: internal/domain/repository.go }
owner: unresolved
updated: 2026-10-16
---

# Architecture

## Shape
- One Go process starts gRPC, an HTTP gateway connected to the local gRPC endpoint, and an optional Kafka consumer.

## Main Components
- `cmd/server`: config and migration setup, `app.New`, and go-core `server.Run`.
- `internal/app`: dependency construction, service registration, gRPC/gateway/consumer wiring.
- `internal/domain`: standard-library-only types and pure rules for transactions, statuses, products, events, catalog, and repository contracts.
- `internal/service`: status-event preparation and persistence flow, history reads, insert-only backfill, and DB-backed message provider.
- `internal/repository`: SQL Server transactions, locking, keyset reads, message reads, and static catalog implementation.
- `internal/handler/kafka`: CloudEvents and JSON decoding, service invocation, and invalid-event DLQ.
- `internal/handler/grpc`: customer APIs, backoffice search and audit, identity, request parsing, and mapping.
- `internal/utils`: standard-library-only PII, decimal, and timezone helpers; no internal imports.
- `proto` and `gen/go`: source and generated API bindings.

## Dependency Direction
- `handler → service → domain ← repository`; `app` wires dependencies and `utils` imports nothing internal.
- SQL belongs in `repository`; pure rules, lists, and constants in `domain`; workflows in `service`; request/JSON/proto handling in `handler`; generic pure helpers in `utils`; environment and wiring in `app`.
- `depguard` enforces standard-library-only `domain` and `utils`, the service allowlist, and the repository/handler dependency bans.

## Runtime Assembly
- The error-message provider loads active definitions at startup and refreshes every 60 seconds.
- Kafka consumption starts only when enabled and all event-consumer settings and publisher are present.
- HTTP customer requests pass through gRPC interceptors; the backoffice service is not registered on the HTTP gateway.

## Persistence Flow
- Event ID check, row lock, snapshot application, optional detail write, and status-event insertion share a SQL transaction.
- Higher status versions replace the snapshot; stale versions do not.
