---
id: knowledge.unknowns
title: Unknowns Ledger
type: knowledge
status: confirmed
confidence: high
source: human
owner: TBD
updated: 2026-05-20
---

# Unknowns

Acknowledged knowledge gaps. Mandatory destination when AI encounters incomplete information. **Guessing forbidden.**

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Ledger — append-only |
| AI writable | Yes — AI **must** write here when encountering a gap, not guess |
| Human confirmation | Not required to add; required to close (resolve) |
| Populated | Throughout lifecycle, especially during init & when agents encounter uncertainty |

## Rules

- Each entry has owner & status.
- Resolution: answer goes to correct semantic location (`01-core/`/`layers/`/`systems/` if human fact, or `inferred.md` if AI inference), then unknown entry marked `resolved`.

## Entries

| ID | Question / Gap | Owner | Created | Status | Resolution |
|---|---|---|---|---|---|
| U-001 | Who is the owner team for this service? `owner: TBD` placed across all files. | TBD | 2026-05-20 | unknown | — |
| U-002 | Does this repo own deployment manifests / CI pipelines / IaC, or are those in another repo? | TBD | 2026-05-20 | unknown | — |
| U-003 | What is the integration test strategy? Only unit tests with sqlmock observed. Any e2e or integration tests against real SQL Server? | TBD | 2026-05-20 | unknown | — |
| U-004 | Is `gen/` policy "always commit generated proto" official, or convention? Should be ADR'd. | TBD | 2026-05-20 | unknown | — |
| U-005 | What is the staleness/migration plan for legacy `.ai/` folder vs new `.forge/context/`? | TBD | 2026-05-20 | unknown | — |
| U-006 | What are the SLA / performance targets for `GetUserHistory` and `CreateTransactionHistory`? Performance budget not in `constraints.md`. | TBD | 2026-05-20 | unknown | — |
| U-007 | Compliance regime? (PII handling, retention policy, audit requirements for transaction history.) | TBD | 2026-05-20 | unknown | — |
| U-008 | Cursor-based pagination roadmap? Current `GetUserHistory` uses numeric offset placeholder. | TBD | 2026-05-20 | unknown | — |
| U-009 | Status lifecycle outside create flow — what is the intended design? README mentions it as "not yet complete." | TBD | 2026-05-20 | unknown | — |
| U-010 | Kafka publisher: when is it enabled, what topics are published, what consumers expect what schema? | TBD | 2026-05-20 | unknown | — |
| U-011 | Redis cache: what is cached, what TTL, what invalidation strategy? | TBD | 2026-05-20 | unknown | — |
