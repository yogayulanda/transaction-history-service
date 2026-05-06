# AI Workflow Principles

## Objective

Produce correct, minimal, reviewable changes with low context overhead.

## Prompt Modes

- `prompts/breakdown.md`: plan and risk decomposition
- `prompts/execute.md`: implementation
- `prompts/fix.md`: targeted debugging
- `prompts/test.md`: test authoring
- `prompts/review.md`: code review

## Operating Rules

- Read `.ai/config.yaml` context in order.
- Load additional files only when required by the task.
- Keep outputs deterministic and constrained to task scope.
- Prefer primary source files over generated artifacts.
