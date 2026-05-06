# Prompt: Add New Feature

> **When to use:** Adding a new API method, filter, or business capability to this service.

---

```
You are a senior Go engineer adding a new feature to transaction-history-service.

Fintech historical store. New features must not break existing API or data contracts.
Layering: handler → service → repository. Go 1.24. Framework: go-core.

Read: .ai/context.md, .ai/conventions.md

== FEATURE ==
{{ Describe the new feature }}

== SCOPE CHECK — answer before implementing ==
1. Is this a query/ingest concern for transaction history? (not processing, not aggregation)
2. Does it need a proto change? If yes, list new RPC or fields.
3. Does it need a schema change? If yes, describe new table or column.
4. Does it need a new error code? If yes, follow TRH-CATEGORY-NUMBER.

Any scope outside historical store → explain which service owns it. Stop here.

== IMPLEMENTATION ==

1. PROTO CHANGE (if needed)
   - New RPC or message fields in proto/history/v1/history.proto
   - Run `make proto` to regenerate gen/

2. DOMAIN
   - New model or interface method in internal/domain/

3. HANDLER
   - Transport validation: nil check, required fields, enum mapping
   - Map proto → domain, call service, map error via coreerrors.ToGRPC

4. SERVICE
   - Business rules, input sanitization, error shaping
   - New error type in service/errors.go if needed

5. REPOSITORY
   - SQL query using dbtx.FromContext
   - Wrap writes in dbtx.WithTx

6. MIGRATION (if schema changes)
   - New file: migrations/transaction/<next_number>_<description>.up.sql
   - Include -- +goose Up and -- +goose Down blocks
   - If new error codes: INSERT into transaction_error_definitions

7. TESTS
   - Handler, service, repository each tested independently

8. DOCS TO UPDATE
   - .ai/context.md if new source file or flow changes
   - README.md API section if new endpoint
```
