---
id: knowledge.confirmations
title: Confirmations Audit Log
type: knowledge
status: confirmed
confidence: high
source: human
owner: unresolved
updated: 2026-05-20
---

# Confirmations

Audit log for status promotions to `confirmed`. Source of truth for when & by whom context was promoted.

## File Meta

| Attribute | Value |
|---|---|
| Source of truth | Audit ledger — append-only, authoritative |
| AI writable | No — only humans write confirmations (AI proposes only) |
| Human confirmation | This IS the confirmation file |
| Populated | Each time an entry is promoted from `inferred`/`assumption` to `confirmed` |

## Rules

- Entries never deleted, only appended.
- Each entry references target file/ID & supporting evidence.

## Entries

| Date | Target ID | From → To | Confirmer | Evidence | Notes |
|---|---|---|---|---|---|
| 2026-05-20 | knowledge.decision.adr-0001 | proposed → accepted | unresolved | `.forge/forge.config.yaml` committed | ADR-0001 forge-context adoption |
| 2026-05-20 | knowledge.unknowns / U-013 | unknown → resolved | unresolved | Team convention override recorded in this confirmation entry | Dominant context language: English (deviates from the default that prefers repo-native Indonesian) |
