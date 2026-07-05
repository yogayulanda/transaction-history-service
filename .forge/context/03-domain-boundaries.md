---
id: system.domain-boundaries
title: Domain Boundaries
type: system
system_type: service
status: confirmed
confidence: high
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: code, ref: internal/domain/transaction.go }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
owner: unresolved
updated: 2026-07-06
---

# Domain Boundaries

## Owned by This Service
- Canonical storage of transaction-history records and product-specific metadata.
- Read APIs for user history and transaction-history detail.
- Fallback/manual ingestion path for new history records.
- Service-local error definition lookup used to shape user-facing messages for this service's error taxonomy.

## Explicitly Not Owned
- Financial transaction execution or settlement.
- Aggregate business reporting logic.
- Upstream source-of-truth validation for producer systems beyond input validation and allowlist checks.
- Authentication/signature implementation details, which are delegated to `go-core` runtime configuration.

## Boundary Conditions
- `reference_id` is treated as a globally unique business identifier across producers.
- `source_service` and `channel` are persisted as first-class business dimensions.
- Kafka ingestion is restricted to a hardcoded producer allowlist in code.
