---
id: system.service-overview
title: Service Overview
type: system
system_type: service
status: confirmed
confidence: high
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: code, ref: cmd/server/main.go }
owner: unresolved
updated: 2026-07-06
---

# Service Overview

## Purpose
- `transaction-history-service` is the service-owned historical store for transaction records used by app history views, downstream reporting consumers, fallback/manual ingestion, and early integration testing.

## Responsibilities
- Store new transaction history records through `CreateTransactionHistory`.
- Serve user transaction history lists through `GetUserHistory`.
- Serve single-record detail through `GetTransactionHistoryDetail`.
- Expose the same contract over gRPC and grpc-gateway HTTP.
- Optionally ingest `transaction.created` Kafka events into the same store.

## Non-Goals
- It is not the transaction processor.
- It is not the reporting engine.
- It does not replace the main event pipeline; the create API remains a fallback/manual path.

## Primary Sources and Consumers
- Producer/source services called out in repository docs: `trxFinance`, `ms-liquiditas`, and agent-payment purchase flows.
- Code-level Kafka producer allowlist currently accepts `trxFinance`, `ms-liquiditas`, and `ms-agent-payment-purchase`.
- Consumers are app-facing history clients and downstream reporting readers.

## Entry Surface
- Entrypoint: `cmd/server/main.go`
- Service wiring: `internal/app/app.go`
- Public contract: `proto/history/v1/history.proto`
- HTTP reference doc: `docs/api.md`
