---
id: knowledge.assumptions
title: Assumptions Ledger
type: knowledge
status: confirmed
confidence: high
source: human
owner: TBD
updated: 2026-05-20
---

# Assumptions

Temporary assumptions ledger. Not a basis for final decisions.

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Ledger — append-only |
| AI writable | Yes — AI adds new assumptions here, **never** to `source: human` files |
| Human confirmation | Not required to add; **required** to promote to `inferred`/`confirmed` |
| Populated | Throughout project lifecycle, especially during planning & init |

## Rules

- Each assumption has owner & status.
- Validated → promote to `inferred.md` (with evidence) → `confirmed` (with entry in `confirmations.md`).
- Invalidated → move to `unknowns.md` or mark `deprecated`.

## Entries

| ID | Assumption | Owner | Created | Status | Notes |
|---|---|---|---|---|---|
| A-001 | Existing `.ai/` folder content is accurate as of 2026-05-20 (used as input for inference) | TBD | 2026-05-20 | assumption | Should be re-confirmed with code review |
| A-002 | Tier `standard` is the right fit; this single-service repo doesn't need `advanced` | TBD | 2026-05-20 | assumption | Re-evaluate if observability/security layers needed |
| A-003 | `infrastructure` layer ownership is partial (migrations + build only); deployment lives elsewhere | TBD | 2026-05-20 | assumption | See U-002 |
| A-004 | `default_mode: implementation` matches the dominant work pattern | TBD | 2026-05-20 | assumption | Adjust if planning/review work dominates |
