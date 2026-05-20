---
id: meta.conventions
title: Conventions & AI Operational Contract
type: meta
status: confirmed
confidence: high
source: human
evidence:
  - { type: doc, ref: ../../../FORGE-CONTEXT-ARCHITECTURE.md }
owner: forge-context-engine
updated: 2026-05-20
---

# Context System Conventions

Rules for **managing the context system itself**. Not product engineering principles (→ `01-core/principles.md`).

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Normative. Always loaded. |
| AI writable | No |
| Human confirmation | Required for any change |
| Populated | Final in runtime; minor adaptation during target repo init |

## Naming & IDs

- Files: `kebab-case.md`
- ID format: `<zone>.<name>` (e.g. `system.payment-service`, `layer.backend`)
- ADR: `ADR-NNNN-title.md`, append-only

## Front-Matter Schema (Required on Every Context File)

```yaml
---
id: <zone>.<name>
title: <human-readable title>
type: meta|core|layer|system|knowledge|mode|generated
system_type: service|app|worker|library|infra-module|platform-component  # type=system only
status: confirmed|inferred|assumption|unknown|deprecated
confidence: high|medium|low
source: human|ai|hybrid
evidence:
  - { type: code|doc|adr|human|external, ref: <path|url> }
owner: <team-or-ref>
updated: YYYY-MM-DD
review_by: YYYY-MM-DD  # optional
---
```

## Status Vocabulary

| Status | Meaning | Authoritative |
|---|---|---|
| `confirmed` | Verified; safe as decision basis | Yes |
| `inferred` | AI-derived with evidence; non-authoritative | No |
| `assumption` | Temporary; not a final decision basis | No |
| `unknown` | Acknowledged gap; **guessing forbidden** | — |
| `deprecated` | No longer applies; not loaded | — |

## Always-Loaded Core

`00-meta/*` + `01-core/*`. Modes **never** re-list core — delta only.

## AI Operational Contract (Normative)

1. AI does not self-promote status — **propose only**.
2. AI does not present `inferred`/`assumption` as fact.
3. On encountering `unknown`, AI stops & asks or records it — **guessing forbidden**.
4. New inferences go to `knowledge/inferred.md` or `generated/`, never to `source: human` files.
5. Without `evidence`, max status is `assumption`.
6. AI does not fabricate architecture, APIs, services, databases, integrations, ownership, or business rules.
7. Treat legacy AI artifacts (`.ai/`, `.claude/`, `AGENTS.md`, etc.) as **reference**, not source-of-truth. Repo code wins on conflict.
8. Tag every `unknowns.md` entry with priority: `blocking` · `important` · `informational`.
9. Use `owner: unresolved` (not `TBD`) when owner is undetermined; create one root unknown `U-OWN`.
10. **Evidence Consistency** — before finalizing context, cross-check critical claims against repo evidence (tables, migrations, entities, repositories, APIs/handlers, workers, integrations, validation rules). If repo shows N, context says N.
11. **Drift Detection** — when repo evidence changes after context was written, mark affected entries as stale, refresh from current code, log unresolved ambiguity in `unknowns.md`.
12. **No Phantom ADRs** — never list `ADR-NNNN` in `architecture.md` (or anywhere as cited evidence) unless the file actually exists. Planned ADRs go to `assumptions.md` or `unknowns.md`.
13. **Implicit Constraint Extraction** — during init, scan code for implicit constraints (enum values, validation rules, required fields, ID semantics, currency/amount rules, status fields, retry/idempotency). Global rules → `constraints.md`. System-specific → `systems/<name>/system.md`. Ambiguous → `unknowns.md`. Weak inference → `inferred.md`.
14. **Validation Semantics** — preserve enforcement layer (service / handler / DB / repository fallback / business intent). Never flatten everything into "required fields". A field DB-constrained but not service-required must be documented as DB-constrained, not service-required.
15. **Internal Table Hygiene** — markdown table cells follow the same conventions as front-matter. Owner cells use `unresolved`, never `TBD`. Status/priority cells use the canonical vocabulary.
16. **Language consistency** — one dominant natural language per repo (chosen at init, override documented). Never translate identifiers (table names, enum values, RPC names, etc.). No mixed-language sentences in narrative content.
17. **Reference stability** — prefer `id`/file references over translated heading text. Citing `core.product` is stable; citing `"Data Sources" section` is fragile.
18. **Runtime vs Seed Semantics** — never describe migration-seeded, bootstrap-only, lookup, or static configuration tables as part of runtime operational write flows unless runtime code actually writes to them. Classify each table by role: `operational-write`, `transactional-write`, `read-only-runtime`, `migration-seeded`, `lookup/reference`, `generated/internal`, or `unknown`.

## Status Promotion

```
assumption ──(evidence)──► inferred ──(human confirmation)──► confirmed
```

Promotion to `confirmed` requires entry in `knowledge/confirmations.md`.

## Lifecycle & Staleness

| Zone | Lifecycle |
|---|---|
| `temp/*` | Single session → deleted (gitignored, never authoritative) |
| `generated/*` | Until regenerated → overwritten. Commit only if stable & useful. Never source-of-truth. Must remain reproducible. |
| `inferred`/`assumption`/`unknown` | Until resolved |
| `core`/`layer`/`system` | Maintained → `deprecated` |
| ADR | Permanent → `superseded`, never deleted |

- `updated` exceeding `governance.staleness_days` → triggers re-review.
- Code change at `evidence` path demotes `confirmed` → `inferred`.

## Anti-Duplication

- One fact, one home.
- Shared context referenced via `id`, **never copied**.
- `systems/*` does not copy `01-core/` or `layers/*` standards.
- `modes/*` does not list `00-meta/*` or `01-core/*`.

## Ownership Rule

Avoid noise from repeated `owner: TBD` placeholders.

| Situation | Action |
|---|---|
| Owner known | Set on every file as canonical team/individual reference |
| Owner unknown | Use `owner: unresolved` and create **one** unknown entry (`U-OWN`) in `knowledge/unknowns.md` |
| Multiple ownership | Use a short ref token (e.g. `team.payments`) and define it once in `glossary.md` |

`owner: TBD` is deprecated. Use `unresolved` (single root unknown) or a real ref.

## Layer Activation Rule

A layer is **activated** only when concrete evidence exists in the target repo.

| Layer | Evidence Required |
|---|---|
| `backend` | Application code (server, API, business logic) |
| `frontend` | UI/web client code |
| `mobile` | iOS/Android/cross-platform code |
| `infrastructure` | IaC (Terraform/Helm/K8s), Dockerfiles for deployment, CI/CD deploy logic |
| `testing` | Test files or test runner configuration |

If evidence is **weak or partial**: activate with `confidence: medium/low` + add unknown entries.
If evidence is **absent**: remove the layer folder and from `forge.config.yaml` → `layers_enabled`.

## README vs Layer Content Policy

| File | Role | Content |
|---|---|---|
| `layers/<x>/README.md` | Entrypoint & TOC | Purpose, navigation links, activation rule. Stays lightweight. |
| `layers/<x>/<x>.md` | Engineering knowledge | Conventions, patterns, tech stack, layer-specific rules. |

README must NOT duplicate `<x>.md` content.

## Unknown Priority Classification

| Priority | Meaning |
|---|---|
| `blocking` | Init or work cannot proceed without resolution |
| `important` | Should be resolved within current sprint/cycle |
| `informational` | Nice to know; resolve when convenient |

## Local Override — Dominant Context Language

**Decision:** English (recorded 2026-05-20).

**Rationale:** The repo's narrative artifacts are mixed (README in Indonesian, code comments in English, error user-messages in Indonesian, AI workflow docs in English). The team chose **English** for `.forge/context/` narrative as the canonical engineering-AI working language.

**Deviation from default rule:** v0.2.1 Language Consistency Rule prefers repo-native language (Indonesian) when one dominates. This repo overrides that preference based on team convention.

### Local Application

Applies across:
`01-core/` · `layers/` · `systems/` · `knowledge/` · `00-meta/glossary.md` (header notes & narrative)

### Identifier Rule (Unchanged)

All technical identifiers stay verbatim regardless of narrative language:
- Table names, column names, enum values, RPC names
- Error codes (`TRH-VAL-001`, etc.)
- Migration filenames
- Repo-native business terms when no equivalent exists (e.g. `bifast`, `rtol`)
- Indonesian error user-messages stored in `transaction_error_definitions` are data, not narrative — they stay verbatim.

### Tracked As

`knowledge/confirmations.md` records this decision; `knowledge/unknowns.md` U-013 is closed by this override.

## Validation Semantics Rule

Validation lives in multiple layers. Context must preserve where each rule is enforced.

| Layer | Source in this repo |
|---|---|
| Service | `internal/service/transaction_service.go` → `sanitizeCreateInput` empty-checks |
| Handler / API | `internal/handler/grpc/handler.go` validators |
| Database | `migrations/transaction/*.sql` `CHECK`, `NOT NULL`, `UNIQUE`, FK |
| Repository | `internal/repository/transaction_sql.go` defaults/fallbacks |
| Business intent | ADRs, `01-core/product.md` |

Never flatten different validation realities into one "required" list. A field DB-constrained but not service-required must be documented as DB-constrained, not service-required.

## Runtime vs Seed Semantics Rule

Table role classification for this repo (see `knowledge/inferred.md` I-017):

| Table | Role |
|---|---|
| `transaction_histories` | `operational-write` · `transactional-write` |
| `transaction_history_details` | `operational-write` · `transactional-write` |
| `transaction_history_status_events` | `operational-write` · `transactional-write` |
| `transaction_error_definitions` | `migration-seeded` · `lookup/reference` |

Create-flow transaction boundary: first 3 tables only. `transaction_error_definitions` is seeded by migration and read at runtime — not written by runtime code.

## Glossary Signal Rule

If all glossary entries share the same `status`/`source`, declare once as a header note and omit from rows. See `00-meta/glossary.md` for applied example.
