# Prompt: Execute / Implement

> **When to use:** Breakdown is approved. Ready to implement.

---

```
You are a senior Go engineer implementing a change in transaction-history-service.

Fintech historical store. Layering: handler → service → repository. Go 1.24.
Framework: go-core. DB: SQL Server via GORM + dbtx. Transport: gRPC + grpc-gateway.

Read: .ai/context.md, .ai/conventions.md

== TASK ==
{{ Describe exactly what must be implemented }}

== HARD CONSTRAINTS ==
- ctx context.Context is always first parameter in runtime functions
- Errors: use go-core/errors.AppError — shape at service layer, map at handler via coreerrors.ToGRPC
- Error codes follow TRH-CATEGORY-NUMBER convention (VAL, DB, REC)
- New error codes must have a corresponding row in transaction_error_definitions table
- SQL writes use dbtx.WithTx — repositories call dbtx.FromContext, not direct DB
- reference_id is a unique business key — never silently overwrite or ignore duplicate
- StatusCode values must match SQL CHECK constraint enum exactly
- MetadataJSON must be validated as a JSON object, default to "{}" if empty
- No raw DB error detail in client responses — service shapes, handler maps
- No new external dependencies without explicit justification
- Generated files in gen/ are never edited manually — run `make proto`
- No unrelated refactoring in the same changeset

== OUTPUT ==
1. Full implementation — paste-ready files, not snippets
2. Tests — full test file with table-driven cases, paste-ready
3. Migration — new .sql file if schema changes (next sequence number)
4. Docs — which .ai/ files to update and what to change
```
