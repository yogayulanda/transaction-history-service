---
id: system.service-overview
title: Service Overview
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: doc, ref: README.md }
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: code, ref: cmd/server/main.go }
owner: unresolved
updated: 2026-10-16
---

# Service Overview

## Purpose
- `transaction-history-service` stores transaction snapshots and ordered status history for customer history views and controlled support searches.

## Responsibilities
- Consume validated transaction status snapshots from Kafka.
- Store decimal-string money, masked destination/source values, product detail, and append-only status events.
- Serve owner-scoped list and detail through gRPC and the HTTP gateway.
- Serve role-gated support search through a separate gRPC-only service.
- Provide an insert-only backfill framework without a public backfill command.

## Non-Goals
- It is not the transaction processor or reporting engine.
- It does not perform arithmetic on transaction money.
- It does not expose user identity in customer request parameters.

## Source and Consumer Policy
- Kafka producer names and topic are environment-configured; public examples use fictional values.
- Customer owner identity comes from authenticated claims.
- Backoffice search requires the configured role and emits an audit event.

## Entry Surface
- Entrypoint: `cmd/server/main.go`
- Service wiring: `internal/app/app.go`
- Public API: `proto/history/v1/history.proto`
- Backoffice API: `proto/history/v1/backoffice.proto`
- Event contract: `docs/events/transaction-status-v1.md`
- HTTP reference: `docs/api.md`
