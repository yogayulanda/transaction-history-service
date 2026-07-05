---
id: system.security-access
title: Security and Access
type: system
system_type: service
status: confirmed
confidence: medium
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: doc, ref: docs/api.md }
  - { type: code, ref: .env.example }
  - { type: code, ref: cmd/server/main.go }
owner: unresolved
updated: 2026-07-06
---

# Security and Access

## Confirmed Controls
- JWT verification and signature middleware are configured through `go-core`, not custom middleware in this repository.
- Config surface exists for `INTERNAL_JWT_*`, `AUTH_SIGNATURE_*`, and `HTTP_PPROF_ENABLED`.
- Service boot relies on `cfg.Validate()` from `go-core` before startup.

## Data Sensitivity
- The service stores end-user identifiers, business reference identifiers, product metadata JSON, and optional error payload details.
- `metadata_json` is stored as opaque JSON object text and is not schema-validated beyond object shape.

## Access Boundary Notes
- No repository-local authorization rules or role model were found in service code.
- Producer authorization for Kafka inbound is currently a static producer-name allowlist, not a cryptographic trust mechanism.

## Caution
- Deployment-time auth requirements, gateway exposure policy, and secret rotation behavior are not documented inside this repo and remain open questions.
