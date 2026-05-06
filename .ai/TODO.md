# AI Context TODO

## Near-Term

- Keep `.ai` aligned whenever API contract, validation behavior, or migration schema changes.
- Keep `README.md` and `docs/api.md` in sync with actual runtime behavior.
- Add context notes when new auth/middleware behavior is introduced from go-core upgrades.

## When Feature Scope Expands

- Add focused context docs only if needed (for example: `api-map.md`, `dependency-map.md`).
- Prefer extending existing canonical files before introducing new files.

## Hygiene

- Do not reintroduce legacy `.ai/context/*` structure.
- Do not include generated files, long logs, or secrets in AI context files.
