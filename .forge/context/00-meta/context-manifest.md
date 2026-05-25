---
id: meta.context-manifest
title: Context Manifest
type: meta
status: confirmed
confidence: high
source: human
evidence:
  - { type: doc, ref: ../../../.forge/forge.config.yaml }
owner: unresolved
updated: 2026-05-25
---

# Context Manifest

Index and routing map for the entire context system. Not a knowledge source.

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | File registry & loading rules |
| AI writable | Yes — propose additions/removals, owner confirms |
| Human confirmation | Required for tier/zone changes |
| Populated | Initialized 2026-05-20 |

## Bootstrap Order

1. `forge.config.yaml`
2. `00-meta/context-manifest.md` ← this file
3. `00-meta/conventions.md`
4. `00-meta/glossary.md` *(if exists)*
5. `01-core/*`
6. `modes/<default_mode>.md` → resolve delta

## Always Loaded

- `forge.config.yaml`
- `00-meta/context-manifest.md`
- `00-meta/conventions.md`
- `00-meta/glossary.md` *(if exists)*
- `01-core/product.md`
- `01-core/architecture.md`
- `01-core/principles.md` *(optional in Minimal tier)*
- `01-core/constraints.md` *(optional in Minimal tier)*

## Selective (Per Mode)

| Zone | Loaded by |
|---|---|
| `layers/<layer>` | Mode referencing that layer |
| `systems/<unit>` | Mode + task intent on that unit |
| `knowledge/decisions/` | `planning`, `implementation`, `execute`, `testing`, `review`, `refactor`; on-demand in `ask`/`incident` |
| `knowledge/assumptions.md`, `unknowns.md` | `planning`, `testing`; on-demand in `ask`/`implementation`/`execute`/`review`/`incident` |
| `knowledge/inferred.md` | `ask`, `implementation`, `execute`, `testing`, `incident`; on-demand in `review`/`refactor` |
| `generated/*` | On-demand |
| `generated/artifacts/*` | Explicit reference, mode handoff, or task relevance only |

## Never Auto-Loaded

- `temp/*` — ephemeral scratch, gitignored.
- Files with `status: deprecated`.

## Validation Rules

- Every file has valid front-matter.
- Every file registered in this manifest.
- Every `id` unique.
- `confirmed`/`inferred` must have `evidence`.
- `source: human` files not written by AI.
- `modes/*` files never list `00-meta/*` or `01-core/*` (delta only).
- Lifecycle artifacts are non-authoritative generated continuity helpers; artifact links never imply workflow, DAG, orchestration, agent memory, or execution-trigger semantics.
- Runtime profile is metadata only; `runtime.non_interactive` remains the controlling interaction flag and automation-safe behavior never implies orchestration, agents, CI/CD, deploy, triggers, or executors.

## File Registry

Format: `path | id | type | status | owner`

```
.forge/forge.config.yaml                                                | (config)                              | config     | n/a       | unresolved
.forge/context/00-meta/context-manifest.md                              | meta.context-manifest                 | meta       | confirmed | unresolved
.forge/context/00-meta/conventions.md                                   | meta.conventions                      | meta       | confirmed | forge-context-engine
.forge/context/00-meta/glossary.md                                      | meta.glossary                         | meta       | inferred  | unresolved
.forge/context/01-core/product.md                                       | core.product                          | core       | inferred  | unresolved
.forge/context/01-core/architecture.md                                  | core.architecture                     | core       | inferred  | unresolved
.forge/context/01-core/principles.md                                    | core.principles                       | core       | inferred  | unresolved
.forge/context/01-core/constraints.md                                   | core.constraints                      | core       | inferred  | unresolved
.forge/context/layers/backend/README.md                                 | layer.backend                         | layer      | unknown   | unresolved
.forge/context/layers/backend/backend.md                                | layer.backend.content                 | layer      | inferred  | unresolved
.forge/context/layers/testing/README.md                                 | layer.testing                         | layer      | unknown   | unresolved
.forge/context/layers/testing/testing.md                                | layer.testing.content                 | layer      | inferred  | unresolved
.forge/context/systems/README.md                                        | systems.readme                        | meta       | unknown   | unresolved
.forge/context/systems/transaction-history-service/system.md            | system.transaction-history-service    | system     | inferred  | unresolved
.forge/context/knowledge/decisions/ADR-0000-template.md                 | knowledge.decision.adr-0000-template  | knowledge  | confirmed | forge-context-engine
.forge/context/knowledge/decisions/ADR-0001-forge-context-adoption.md   | knowledge.decision.adr-0001           | knowledge  | accepted  | unresolved
.forge/context/knowledge/assumptions.md                                 | knowledge.assumptions                 | knowledge  | confirmed | unresolved
.forge/context/knowledge/unknowns.md                                    | knowledge.unknowns                    | knowledge  | confirmed | unresolved
.forge/context/knowledge/inferred.md                                    | knowledge.inferred                    | knowledge  | confirmed | unresolved
.forge/context/knowledge/confirmations.md                               | knowledge.confirmations               | knowledge  | confirmed | unresolved
.forge/context/modes/ask.md                                             | mode.ask                              | mode       | confirmed | forge-context-engine
.forge/context/modes/planning.md                                        | mode.planning                         | mode       | confirmed | forge-context-engine
.forge/context/modes/implementation.md                                  | mode.implementation                   | mode       | confirmed | forge-context-engine
.forge/context/modes/execute.md                                         | mode.execute                          | mode       | confirmed | forge-context-engine
.forge/context/modes/testing.md                                         | mode.testing                          | mode       | confirmed | forge-context-engine
.forge/context/modes/review.md                                          | mode.review                           | mode       | confirmed | forge-context-engine
.forge/context/modes/incident.md                                        | mode.incident                         | mode       | confirmed | forge-context-engine
.forge/context/modes/refactor.md                                        | mode.refactor                         | mode       | confirmed | forge-context-engine
```

## Active Configuration Snapshot

- Tier: `standard`
- Layers enabled: `backend`, `testing` *(infrastructure deactivated per v0.2.2 Layer Activation Rule — no IaC/deploy evidence)*
- Systems registered: `transaction-history-service` (type: `service`)
- Default mode: `implementation`
- Runtime profile: `local`
- Runtime non-interactive: `false`
