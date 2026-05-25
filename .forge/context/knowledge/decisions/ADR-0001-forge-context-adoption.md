---
id: knowledge.decision.adr-0001
title: ADR-0001 — Forge Context Engine Adoption
type: knowledge
status: accepted
source: human
evidence:
  - { type: external, ref: forge-context-engine repo }
  - { type: doc, ref: .forge/forge.config.yaml }
owner: unresolved
updated: 2026-05-20
---

# ADR-0001: Forge Context Engine Adoption

- **Status:** accepted
- **Date:** 2026-05-20
- **Decision makers:** Repo owner (unresolved — see `knowledge/unknowns.md` U-OWN)

## Context

The repo previously used a flat `.ai/` folder with markdown files for AI context (`context.md`, `architecture.md`, `decisions.md`, etc.). While useful, the structure lacked:

- Explicit separation between facts, inferences, assumptions, and unknowns.
- Standardized front-matter and validation rules.
- Modular loading per work mode (planning, implementation, review, testing).
- Anti-hallucination guarantees via evidence requirements.

## Decision

Adopt `forge-context-engine` v0.5 (Standard tier) as the canonical AI context system for this repo. Place the runtime under `.forge/context/`. Keep existing `.ai/` folder as legacy until migration is complete (a future ADR will record the legacy migration plan; tracked as `knowledge/unknowns.md` U-005).

## Alternatives Considered

1. **Stay with flat `.ai/` folder** — rejected: lacks structure for scale, no anti-hallucination guarantees.
2. **Custom in-repo solution** — rejected: reinventing what `forge-context-engine` already specifies.
3. **Use external doc tool (e.g., Notion, Confluence)** — rejected: AI agents work best with in-repo context; tool lock-in.

## Consequences

### Positive

- Clear separation of six knowledge states (confirmed/inferred/assumption/unknown/decision/confirmation).
- Token-efficient selective loading via modes.
- Front-matter validation rules enable future automation.
- Brownfield-safe: code is source of truth; context is cached view.

### Negative

- Migration cost from `.ai/` → `.forge/context/` (incremental, see future ADR).
- Dual context during transition.
- New convention for contributors to learn.

### Trade-offs

- Larger initial structure (~22 files vs ~16). Mitigated by placeholder-light design.

## Evidence

- `.forge/forge.config.yaml` — engine configuration committed.
- `forge-context-engine` repo: https://github.com/yogayulanda/forge-context-engine
