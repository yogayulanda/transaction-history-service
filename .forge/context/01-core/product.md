---
id: core.product
title: Product Context
type: core
status: inferred
confidence: high
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: doc, ref: .ai/context.md }
owner: unresolved
updated: 2026-05-20
---

# Product

## Summary

`transaction-history-service` is a centralized transaction history store. Acts as a *historical store* — not a transaction processor and not a reporting engine.

## Domain & Problem Space

- Domain: financial transaction history persistence.
- Problem: applications and downstream consumers need consistent, queryable transaction history storage that can also serve as a fallback ingestion path when the primary event/Kafka pipeline is unavailable.

## Users & Stakeholders

- Applications (UI transaction history menu)
- Downstream reporting consumers
- Integration teams using this service as fallback/manual ingestion path
- QA teams during early integration testing (insert testing)

## Producers (Data Sources)

- `trxFinance`
- `ms-liquiditas`
- `agent-payment-purchase`

## System Boundaries

### IN Scope

- Store transactions via internal API `CreateTransactionHistory`
- Serve user transaction history list (`GetUserHistory`)
- Serve single transaction detail (`GetTransactionHistoryDetail`)
- Provide historical data ready for consumption by apps and downstream reporting consumers

### OUT of Scope

- Processing financial transactions
- Computing business aggregate reports within this service
- Replacing the primary event/Kafka ingestion path

## Core Product Terms

- **reference_id** — business transaction id; unique across all producers
- **source_service** — producer system that sent the transaction
- **channel** — business dimension identifying the transaction origin (multi-app)
- **product_group**, **product_type** — transaction product taxonomy
- **transaction_route** — routing path identifier for the transaction
- **status_code** — transaction status enum; must stay in sync with SQL constraints
- **metadata_json** — JSON object for product-specific attributes; core fields must NOT be moved here
