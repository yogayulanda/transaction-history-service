---
id: layer.infrastructure.content
title: Infrastructure Layer Conventions
type: layer
status: inferred
confidence: medium
source: ai
evidence:
  - { type: code, ref: migrations/ }
  - { type: code, ref: Makefile }
  - { type: code, ref: .env.example }
  - { type: doc, ref: README.md }
  - { type: doc, ref: .ai/integrations.md }
owner: unresolved
updated: 2026-05-20
---

# Infrastructure Layer

Service-owned infrastructure concerns: migrations, build tooling, runtime configuration.

> Note: deployment manifests, CI/CD pipelines, and cluster config tampaknya hidup di luar repo ini. Confidence `medium` until repo ownership of those concerns is confirmed (see `knowledge/unknowns.md`).

## Database Migrations

- Tool: `goose` (`github.com/pressly/goose/v3`)
- Location: `migrations/transaction/` (production-bound)
- Dev seed: `migrations/dev/dev_seed_transaction_history.sql` (local only)
- Naming: `NNNN_description.up.sql` / `NNNN_description.down.sql`
- Auto-run: controlled by `MIGRATION_AUTO_RUN` env

### Existing Migrations

- `0001_init_transaction_history.up.sql` — baseline schema
- `0002_seed_transaction_history.up.sql` — production no-op (kept for parity)

## Build Tooling

| Tool | Purpose | How |
|---|---|---|
| `protoc` + plugins | Generate gRPC + gateway code | `make proto` |
| `go build` | Compile binary | implicit in `make run` |
| `go test` | Run tests | `make test` |
| `go mod tidy` | Sync deps | manual |

Required protoc plugins:
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- `protoc-gen-grpc-gateway`

## Runtime Configuration

Config keys (loaded via go-core):

| Category | Keys |
|---|---|
| Transport | `GRPC_PORT`, `HTTP_PORT` |
| Database | SQL Server connection (per `transaction_history` DB) |
| Migration | `MIGRATION_AUTO_RUN` |
| Auth | `INTERNAL_JWT_*`, `AUTH_SIGNATURE_*` |
| Debug | `HTTP_PPROF_ENABLED` |

Source of truth for required keys: `.env.example`.

## Generated Code Policy

- `gen/` contains regenerable code.
- Currently committed (per repo policy as of init).
- Regenerate via `make proto` after any `proto/` change.

## Observability Stack

Provided by go-core:
- Prometheus metrics (`/metrics`)
- OpenTelemetry tracing (otlp exporter)
- Structured logging (zap-based)

## Out of Scope (for this layer in this repo)

- Kubernetes manifests
- CI/CD pipeline definitions
- Cloud infrastructure as code (Terraform, etc.)
- Secret management beyond `.env`

These concerns may live in separate repos. Recorded in `knowledge/unknowns.md` for clarification.
