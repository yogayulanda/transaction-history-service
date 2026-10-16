---
id: system.runtime-deployment-config
title: Runtime Deployment and Config
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: .env.example }
  - { type: code, ref: internal/app/kafka_inbound.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-10-16
---

# Runtime Deployment and Config

## Startup
- Load `.env`, configure migrations, validate shared config, run migrations, create go-core app, construct service handlers, register gRPC services, and start gRPC/HTTP plus optional Kafka.
- HTTP gateway calls the local gRPC endpoint with insecure transport credentials; production transport assumptions require deployment validation.
- Outside `APP_ENV=local`, default claims identity requires internal JWT enabled.

## Core Configuration
- Service/group identity: `SERVICE_NAME`.
- SQL Server: `DB_LIST=transaction_history` and `DB_TRANSACTION_HISTORY_*`.
- Ports: `GRPC_PORT`, `HTTP_PORT`.
- Inbound: `KAFKA_INBOUND_ENABLED`, `HISTORY_STATUS_TOPIC`, `HISTORY_ALLOWED_SOURCES`.
- Catalog: `HISTORY_ALLOWED_APPS` (default `app_a,app_b`).
- HMAC: `HISTORY_HASH_KEY` hex-encodes at least 32 bytes; required by app startup.
- Backoffice: `HISTORY_BACKOFFICE_ROLE` (default `backoffice`).
- go-core runtime: v`1.2.1`; Go toolchain `1.26.9`, directive `1.26.0`.

## Developer Commands
- `make generate`, `make test`, `make vet`, `make lint`, `make vuln`, `make fmt-check`, `make check`, `make run`.

## Deployment Unknowns
- This repository has no deployment manifests, environment topology, or rollback procedure.
- Live SQL Server, Kafka, and production authentication configuration still require pre-deploy verification.
