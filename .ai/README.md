# .ai/ — AI Context for transaction-history-service

This folder is the canonical AI context for this repository.

## Read Order

1. `.ai/context.md` — service purpose, runtime flow, hard constraints
2. `.ai/architecture.md` — layer boundaries and system structure
3. Task-specific files (modules, security, transactions, etc.)

## Task Routing

| Working on | Read |
|---|---|
| API behavior / contract | `.ai/modules.md`, `.ai/data-flow.md` |
| SQL persistence / migration | `.ai/transactions.md`, `.ai/modules.md` |
| Auth / JWT / signature / pprof | `.ai/security.md`, `.ai/integrations.md` |
| Runtime config / startup | `.ai/integrations.md`, `.ai/workflow.md` |
| Code style / scope safety | `.ai/conventions.md` |
| Design rationale | `.ai/decisions.md` |

## Structure

```
.ai/
├── context.md
├── architecture.md
├── modules.md
├── data-flow.md
├── transactions.md
├── security.md
├── integrations.md
├── conventions.md
├── decisions.md
├── workflow.md
├── AI_WORKFLOW.md
├── STATUS.md
├── config.yaml
├── prompts/
├── tasks/
└── scripts/
```

## Maintenance Rule

- Update the relevant `.ai/*.md` file whenever behavior changes.
- Keep entries short and factual.
- Do not duplicate generated code or full logs in context files.
