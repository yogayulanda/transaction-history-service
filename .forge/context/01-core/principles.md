---
id: core.principles
title: Engineering Principles
type: core
status: inferred
confidence: high
source: ai
evidence:
  - { type: doc, ref: .ai/conventions.md }
  - { type: doc, ref: .ai/architecture.md }
  - { type: doc, ref: .ai/decisions.md }
owner: unresolved
updated: 2026-05-20
---

# Principles

## Layer Discipline

- Strict separation: `handler → service → repository`. No cross-layer shortcuts.
- No transport logic in service or repository.
- No SQL in handler or service.
- Layers communicate via interfaces (domain repository interface, service contracts).

## Framework Boundary

- Service owns domain behavior; `go-core` owns infrastructure behavior.
- Auth, signature, JWT, observability, transport middleware → framework concerns.
- Never reimplement infrastructure that go-core already provides.

## Change Discipline

- Prefer additive changes; avoid broad refactors.
- Keep changes minimal and scoped to requested behavior.
- Preserve existing package boundaries.
- Add tests with behavior changes.
- Avoid unrelated formatting-only diffs.

## Error Contract

- Service errors built as `coreerrors.AppError`.
- Handler converts via `coreerrors.ToGRPC`.
- Error metadata follows go-core taxonomy: `domain`, `category`, `number`, `finality`, `retryable`.
- Response errors sanitized by go-core contract.

## API Contract

- gRPC-first. Proto is source of truth.
- HTTP API is grpc-gateway projection, not an independent contract.
- API status enums must stay in sync with SQL constraints.

## Persistence

- Multi-table writes for create flow → single DB transaction (`dbtx.WithTx`).
- Field placement rule: core business fields stay as columns; only product-specific attributes go into `metadata_json`.

## Context Hygiene

- Concise summaries over copied code in context files.
- Never include generated file bodies in AI context.
- Never include secret values from `.env`.
