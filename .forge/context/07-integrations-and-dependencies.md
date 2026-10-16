---
id: system.integrations-dependencies
title: Integrations and Dependencies
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: go.mod }
  - { type: code, ref: internal/app/app.go }
  - { type: code, ref: internal/handler/kafka/status_event.go }
owner: unresolved
updated: 2026-10-16
---

# Integrations and Dependencies

## Core Runtime
- Public `github.com/yogayulanda/go-core` v`1.2.1` supplies config, logging, server, migration, auth, and messaging infrastructure.
- SQL Server is accessed through `database/sql`; event writes use SQL Server lock hints.
- Kafka consumer and DLQ publisher are created through go-core messaging.

## Kafka Configuration
- `KAFKA_INBOUND_ENABLED` controls the status consumer.
- `HISTORY_STATUS_TOPIC` sets the topic; DLQ is topic plus `.dlq`.
- Consumer group is `SERVICE_NAME`; default concurrency is 3, retry maximum 5, delay 5 seconds.
- `HISTORY_ALLOWED_SOURCES` is required when enabled.
- CloudEvents attributes are restored from `ce_*` headers to handler context.

## Failure Contract
- Validation failures are manually published to DLQ with `x-dlq-reason`, then acknowledged after publish success.
- DB and failed-DLQ publish errors remain retryable; runtime DLQ failure does not commit the input.
- The DLQ retains the original event payload unchanged so it can be replayed after a consumer fix; the payload must never be written to logs.
- Deployment requirement for the internal fork (I-A): each `<topic>.dlq` must use the source topic's ACL, TLS policy, and retention period.

## Protocols and Tooling
- Customer API: gRPC and grpc-gateway v2; HTTP gateway dials the local gRPC endpoint.
- Backoffice API: gRPC only.
- Proto generation uses `protoc`, Go/gRPC/gateway plugins, and cached Google API proto sources.
