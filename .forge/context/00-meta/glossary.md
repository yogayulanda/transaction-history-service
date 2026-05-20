---
id: meta.glossary
title: Glossary
type: meta
status: inferred
confidence: high
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: doc, ref: .ai/context.md }
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: code, ref: migrations/transaction/0001_init_transaction_history.up.sql }
owner: unresolved
updated: 2026-05-20
---

# Glossary

Canonical definitions for **project/domain-specific** terms ambiguous in this repo.

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Local repo terminology |
| AI writable | Propose via `knowledge/inferred.md` first; promotion requires owner confirmation |
| Human confirmation | Required before entry lands here |
| Populated | During Context Initialization; grows incrementally |

## Defaults

> All entries below: `status: inferred`, `source: ai` — unless a row overrides explicitly.

## Rules

- Only project/domain-specific terms, acronyms, or terms with special meaning in this repo.
- No general engineering terms.
- One line per term. Identifier-shaped terms stay verbatim.

## Entries

| Term | Canonical Definition | Alias |
|---|---|---|
| reference_id | Business transaction id; unique across all producers (UNIQUE index) | — |
| external_ref_id | Optional secondary reference from upstream system | `ExternalRefID` |
| source_service | Producer system that originated the transaction | — |
| channel | Business dimension identifying which app/origin a transaction came from | — |
| product_group | Top-level taxonomy: `ppob`, `transfer`, `cash` | — |
| product_type | Sub-taxonomy under `product_group` | — |
| transaction_route | Routing path: `internal`, `bifast`, `rtol`, `switching`, `partner_h2h` | — |
| direction | Money flow: `debit` or `credit` | — |
| status_code | Transaction status enum (7 values; see `core.constraints`) | — |
| metadata_json | JSON object column for product-specific attributes; not for core fields | — |
| go-core | Internal Go framework providing bootstrap, transport, auth, observability | — |
| AppError | Error contract from go-core (`coreerrors.AppError`) used across service layer | — |
| coreerrors.ToGRPC | Handler-side error converter from `AppError` to gRPC status | — |
| trxFinance | Transaction producer system | — |
| ms-liquiditas | Transaction producer system | `ms-` = microservice |
| agent-payment-purchase | Transaction producer system | — |
| dbtx.WithTx | go-core transaction wrapper used by repository for atomic multi-table writes | — |
| ServiceLog / DBLog / TransactionLog | go-core log flavors for service / DB / transaction events | — |
| TRH-VAL-001 / TRH-VAL-002 / TRH-DB-001 / TRH-REC-001 | Pre-defined user-facing error codes (see `core.constraints`) | — |
