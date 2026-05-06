# Conventions

## Code Change Rules

- Keep changes minimal and scoped to the requested behavior.
- Preserve existing package boundaries.
- Add tests with behavior changes.
- Avoid unrelated formatting-only diffs.

## Layer Rules

- Handler: transport validation and mapping only.
- Service: business rules and error shaping.
- Repository: SQL and transaction orchestration only.

## Context Hygiene

- Prefer concise summaries over long copied code.
- Never include generated file bodies in AI context.
- Never include secret values from `.env`.
