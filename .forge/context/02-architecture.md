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
  - { type: code, ref: internal/repository/history_sql.go }
owner: unresolved
updated: 2026-10-16
---

# Architecture

## Shape
- One Go process starts gRPC, an HTTP gateway connected to the local gRPC endpoint, and an optional Kafka consumer.

## Main Components
- `cmd/server`: config, migrations, service registrations, gateway endpoint, runtime startup.
- `internal/app`: SQL repositories, message provider, identity resolver, handlers, and Kafka consumer.
- `internal/domain`: transaction snapshot, status event, detail, owner, and persistence contracts.
- `internal/handler/kafka`: CloudEvents/payload validation, PII preparation, and invalid-event DLQ.
- `internal/handler/grpc`: owner APIs and gRPC-only backoffice search.
- `internal/backfill`: source/runner/report framework; no public executable.
- `internal/repository`: SQL Server transactions, locking, keyset reads, and event history.
- `proto` and `gen/go`: source and generated API bindings.

## Runtime Assembly
- The error-message provider loads active definitions at startup and refreshes every 60 seconds.
- Kafka consumption starts only when enabled and all event-consumer settings and publisher are present.
- HTTP customer requests pass through gRPC interceptors; the backoffice service is not registered on the HTTP gateway.

## Persistence Flow
- Event ID check, row lock, snapshot application, optional detail write, and status-event insertion share a SQL transaction.
- Higher status versions replace the snapshot; stale versions do not.
