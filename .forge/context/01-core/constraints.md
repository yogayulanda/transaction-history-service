---
id: core.constraints
title: Hard Constraints
type: core
status: inferred
confidence: high
source: ai
evidence:
  - { type: doc, ref: .ai/security.md }
  - { type: doc, ref: .ai/integrations.md }
  - { type: doc, ref: README.md }
  - { type: code, ref: go.mod }
owner: TBD
updated: 2026-05-20
---

# Constraints

## Required Technology

| Constraint | Source |
|---|---|
| Go 1.24.3+ | `go.mod` |
| `go-core` framework (local replace `../go-core`) | `go.mod` replace directive |
| SQL Server (DB: `transaction_history`) | `.ai/integrations.md`, README |
| gRPC + grpc-gateway | proto + gen layout |

## Auth & Security (Framework-Controlled)

- Internal JWT verification: env-toggled (`INTERNAL_JWT_ENABLED`).
- Signature middleware: env-toggled (`AUTH_SIGNATURE_ENABLED`).
- HTTP pprof endpoint: env-toggled (`HTTP_PPROF_ENABLED`).
- These are security-impacting; never weaken in handler code.

## Service Rules

- Never log secrets or raw auth credentials.
- Keep error responses sanitized via app error contract.
- Do not weaken auth checks in handlers.
- Do not bypass framework middleware with custom transport handling.

## Data Integrity

- `reference_id` must be unique across all producers.
- API status enum values must stay consistent with SQL constraints.
- Multi-table writes for create must occur in a single transaction.

## Field Schema Rule

Core business fields stay as columns (not in `metadata_json`):
- `reference_id`, `source_service`, `channel`, `product_group`, `product_type`, `transaction_route`, `status_code`, `transaction_time`

`metadata_json`:
- Must be a valid JSON object.
- Only for product-specific attributes.

## Build & Tooling

- `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway` required for codegen.
- Generated proto code must be regenerated when `proto/` changes.
- `go.mod` and `go.sum` must remain synchronized (`go mod tidy`).
