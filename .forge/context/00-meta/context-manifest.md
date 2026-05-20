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
updated: 2026-05-20
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
4. `00-meta/glossary.md`
5. `01-core/*`
6. `modes/<default_mode>.md` → resolve delta

## Always Loaded

- `forge.config.yaml`
- `00-meta/context-manifest.md`
- `00-meta/conventions.md`
- `00-meta/glossary.md`
- `01-core/product.md`
- `01-core/architecture.md`
- `01-core/principles.md`
- `01-core/constraints.md`

## Selective (Per Mode)

| Zone | Loaded by |
|---|---|
| `layers/<layer>` | Mode referencing that layer |
| `systems/<unit>` | Mode + task intent on that unit |
| `knowledge/decisions/` | `implementation`, `review` |
| `knowledge/assumptions.md`, `unknowns.md` | `planning`, `testing` |
| `knowledge/inferred.md` | `implementation` |
| `generated/*` | On-demand |

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

## File Registry

Format: `path | id | type | status | owner`

```
.forge/forge.config.yaml                                                | (config)                              | config     | n/a       | TBD
.forge/context/00-meta/context-manifest.md                              | meta.context-manifest                 | meta       | confirmed | TBD
.forge/context/00-meta/conventions.md                                   | meta.conventions                      | meta       | confirmed | forge-context-engine
.forge/context/00-meta/glossary.md                                      | meta.glossary                         | meta       | inferred  | TBD
.forge/context/01-core/product.md                                       | core.product                          | core       | inferred  | TBD
.forge/context/01-core/architecture.md                                  | core.architecture                     | core       | inferred  | TBD
.forge/context/01-core/principles.md                                    | core.principles                       | core       | inferred  | TBD
.forge/context/01-core/constraints.md                                   | core.constraints                      | core       | inferred  | TBD
.forge/context/layers/backend/README.md                                 | layer.backend                         | layer      | unknown   | TBD
.forge/context/layers/backend/backend.md                                | layer.backend.content                 | layer      | inferred  | TBD
.forge/context/layers/infrastructure/README.md                          | layer.infrastructure                  | layer      | unknown   | TBD
.forge/context/layers/infrastructure/infrastructure.md                  | layer.infrastructure.content          | layer      | inferred  | TBD
.forge/context/layers/testing/README.md                                 | layer.testing                         | layer      | unknown   | TBD
.forge/context/layers/testing/testing.md                                | layer.testing.content                 | layer      | inferred  | TBD
.forge/context/systems/README.md                                        | systems.readme                        | meta       | unknown   | TBD
.forge/context/systems/transaction-history-service/system.md            | system.transaction-history-service    | system     | inferred  | TBD
.forge/context/knowledge/decisions/ADR-0000-template.md                 | knowledge.decision.adr-0000-template  | knowledge  | confirmed | forge-context-engine
.forge/context/knowledge/decisions/ADR-0001-forge-context-adoption.md   | knowledge.decision.adr-0001           | knowledge  | accepted  | TBD
.forge/context/knowledge/assumptions.md                                 | knowledge.assumptions                 | knowledge  | confirmed | TBD
.forge/context/knowledge/unknowns.md                                    | knowledge.unknowns                    | knowledge  | confirmed | TBD
.forge/context/knowledge/inferred.md                                    | knowledge.inferred                    | knowledge  | confirmed | TBD
.forge/context/knowledge/confirmations.md                               | knowledge.confirmations               | knowledge  | confirmed | TBD
.forge/context/modes/planning.md                                        | mode.planning                         | mode       | confirmed | forge-context-engine
.forge/context/modes/implementation.md                                  | mode.implementation                   | mode       | confirmed | forge-context-engine
.forge/context/modes/review.md                                          | mode.review                           | mode       | confirmed | forge-context-engine
.forge/context/modes/testing.md                                         | mode.testing                          | mode       | confirmed | forge-context-engine
```

## Active Configuration Snapshot

- Tier: `standard`
- Layers enabled: `backend`, `infrastructure`, `testing`
- Systems registered: `transaction-history-service` (type: `service`)
- Default mode: `implementation`
