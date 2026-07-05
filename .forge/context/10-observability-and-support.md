---
id: knowledge.observability-support
title: Observability and Support
type: knowledge
status: confirmed
confidence: medium
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/service/transaction_service.go }
  - { type: code, ref: internal/repository/transaction_sql.go }
  - { type: code, ref: .env.example }
owner: unresolved
updated: 2026-07-06
---

# Observability and Support

## Logging
- Service layer emits service logs for create, list, and detail operations.
- Service layer emits transaction logs for create outcomes.
- Repository emits DB logs with DB name `transaction_history`.
- Kafka inbound emits service logs with event metadata, status, and DLQ outcome.
- Startup logs enumerate derived HTTP endpoints, gRPC methods, and configured ports.

## Runtime Endpoints and Signals
- Repository docs state `go-core` exposes `health`, `ready`, `version`, and `metrics` endpoints.
- `coreserver.LogStartupReadiness` is used during startup with a 10-second readiness observation window.
- OTLP endpoint and trace sampling config are present in `.env.example`.
- `HTTP_PPROF_ENABLED` exists as a runtime toggle.

## Current Gaps
- No repository-local alert rules, dashboards, SLOs, or support playbooks were found.
