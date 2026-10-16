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
updated: 2026-10-15
---

# Runtime Deployment and Config

## Startup Sequence
- Load `.env`
- Set migration auto-run, database name, and directory in code
- Validate config
- Auto-run migrations from `migrations/transaction`
- Build `go-core` app
- Build service app wiring
- Start gRPC server and HTTP gateway
- Start Kafka inbound only when enabled and available

## Core Config Surface
- Service identity: `SERVICE_NAME=transaction-history-service`
- Ports: `GRPC_PORT`, `HTTP_PORT`
- SQL Server connection: `DB_TRANSACTION_HISTORY_*`
- Observability: `APM_ENDPOINT`, `METRICS_ENABLED`, `DIAG_ENABLED`, `HTTP_PPROF_ENABLED`
- Signature validation: a non-empty `AUTH_SIGNATURE_MASTER_KEY`; internal JWT remains disabled.
- Kafka: non-empty `KAFKA_BROKERS` enables Kafka; `KAFKA_INBOUND_ENABLED` controls inbound consumption and malformed non-empty values fail startup.
- Runtime versions: Go `1.26.9` toolchain, module directive `1.26.0`, public `go-core v1.2.1`.

## Developer Commands
- `make proto` / `make generate`
- `make test` (race-enabled)
- `make vet`, `make lint`, `make vuln`, `make fmt-check`, `make check`
- `make run`

## Deployment Unknowns
- This repo does not include deployment manifests, CI/CD pipelines, or environment-specific topology docs.
- SQL Server migration and HTTP smoke checks are pre-deploy gates; see `99-open-questions.md`.
