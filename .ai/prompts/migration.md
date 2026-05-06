# Prompt: Write Migration

> **When to use:** Schema change, new table, new index, or new error_definition rows needed.

---

```
You are a senior Go engineer writing a SQL Server migration for transaction-history-service.

DB: SQL Server (dbo schema). Migration tool: Goose.
All migrations are in migrations/transaction/ and numbered sequentially.
Never modify existing migration files — always add a new one.

Read: .ai/context.md, migrations/transaction/ (existing files for context)

== CHANGE ==
{{ Describe what schema change is needed and why }}

== RULES ==
- File name: <next_number>_<snake_case_description>.up.sql
- Always include -- +goose Up and -- +goose Down blocks
- Down block must be idempotent: use IF OBJECT_ID / IF EXISTS guards
- New columns must have DEFAULT or be NULLable — no silent breakage on existing rows
- New CHECK constraints must match domain enum values exactly
- New indexes: name pattern IX_<table>_<columns>, unique: UX_<table>_<columns>
- FK constraints: ON DELETE CASCADE only when child rows have no independent lifecycle
- New error codes: INSERT into transaction_error_definitions with TRH-CATEGORY-NUMBER format

== DELIVER ==
1. Full migration file — paste-ready, with Up and Down
2. Impact summary — which existing queries or constraints are affected
3. Rollback safety — confirm Down block is safe to run after Up
```
