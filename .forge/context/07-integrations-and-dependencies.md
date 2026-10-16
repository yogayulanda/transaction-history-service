---
id: system.integrations-dependencies
title: Integrations and Dependencies
type: system
system_type: service
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: go.mod }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
owner: unresolved
updated: 2026-10-15
---

# Integrations and Dependencies

## Core Runtime Dependency
- The service depends on the public `github.com/yogayulanda/go-core` module at `v1.2.1`; `go.mod` has no local sibling `replace`.
- `go-core` provides config loading, app bootstrapping, server startup, gateway/grpc servers, migration execution, logging, cache, DB transaction helpers, and messaging abstractions.

## Infrastructure Dependencies
- SQL Server via `gorm.io/driver/sqlserver` and `gorm.io/gorm`.
- Optional Redis cache client exposed by `go-core`; current service logic wires it but does not actively use it.
- Optional Kafka publisher and consumer via `go-core/messaging`.

## Messaging Contracts
- Inbound topic: `transaction-history.transaction.created`
- DLQ topic: `transaction-history.transaction.created.dlq`
- Consumer group: `transaction-history-service`
- Default consumer concurrency: `3`
- Default retry config: `5` attempts with `5s` retry delay
- Consumer concurrency is `3`; go-core v1.2.1 assigns each Kafka partition to a deterministic worker and preserves processing/commit order within that partition.
- Service-level non-retryable events are manually published with `x-original-topic` and `x-error-code` headers.
- Errors that exhaust consumer retries and are runtime-DLQ-published carry `x-dlq-reason` and `x-original-topic` headers. Failed DLQ publication does not commit the original message and blocks further processing for that partition.

## Protocol Dependencies
- gRPC and grpc-gateway v2
- Google API annotations for HTTP mapping
- OpenAPI file under `docs/openapi.yaml` as a secondary reference surface
