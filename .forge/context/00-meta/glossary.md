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
owner: TBD
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

## Rules

- Only project/domain-specific terms, acronyms, or terms with special meaning in this repo.
- No general engineering terms.
- One line per term.

## Entries

| Term | Canonical Definition | Status | Alias |
|---|---|---|---|
| reference_id | Business transaction id; unique across all producers | inferred | — |
| source_service | Producer system that originated the transaction | inferred | — |
| channel | Business dimension identifying which app/origin a transaction came from | inferred | — |
| product_group | Top-level taxonomy of transaction product | inferred | — |
| product_type | Sub-taxonomy under product_group | inferred | — |
| transaction_route | Routing path identifier for the transaction | inferred | — |
| status_code | Transaction status enum; sync with SQL constraints | inferred | — |
| metadata_json | JSON object column for product-specific attributes; not for core fields | inferred | — |
| go-core | Internal Go framework providing bootstrap, transport, auth, observability | inferred | — |
| AppError | Error contract from go-core (`coreerrors.AppError`) used across service layer | inferred | — |
| coreerrors.ToGRPC | Handler-side error converter from AppError to gRPC status | inferred | — |
| trxFinance | Transaction producer system | inferred | — |
| ms-liquiditas | Transaction producer system (microservice naming convention) | inferred | ms = microservice |
| agent-payment-purchase | Transaction producer system | inferred | — |
| dbtx.WithTx | go-core transaction wrapper used by repository for atomic multi-table writes | inferred | — |
| ServiceLog / DBLog / TransactionLog | go-core log flavors for service / DB / transaction events | inferred | — |
