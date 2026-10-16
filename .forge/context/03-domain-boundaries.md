---
id: system.domain-boundaries
title: Domain Boundaries
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: internal/domain/transaction.go }
  - { type: code, ref: internal/handler/kafka/status_event.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-10-16
---

# Domain Boundaries

## Owned
- Transaction snapshot, product-specific detail, and append-only status history.
- Validation and ordering of incoming status snapshots.
- Owner-scoped customer list/detail and role-gated support search.
- Insert-only backfill primitives for source adapters supplied by a fork.
- Masking and hashing of sensitive account identifiers before persistence.

## Not Owned
- Transaction execution, settlement, or financial calculations.
- Producer-system source-of-truth validation beyond this service's catalog and payload rules.
- A public backfill command or producer-specific legacy adapter.

## Boundary Rules
- Customer identity is derived from authenticated claims, never customer request fields.
- Producer source names are supplied through configuration; public examples use fictional values.
- Money is carried as decimal strings and is not recalculated by this service.
- Backoffice search is gRPC-only and audited.
