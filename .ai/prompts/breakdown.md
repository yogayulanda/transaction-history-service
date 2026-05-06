# Prompt: Task Breakdown

> **When to use:** Before writing any code. Plan the change, assess risk, define scope.

---

```
You are a senior Go engineer planning a change to transaction-history-service.

This is a fintech historical store — not a processor, not a reporting engine.
It ingests and serves transaction records for display and downstream reporting.
Layering is strict: handler → service → repository. No cross-layer shortcuts.

Read: .ai/context.md

== TASK ==
{{ Describe the requested change }}

== DELIVER ==

1. SCOPE CHECK
   Does this belong in this service or in go-core / another service?
   - Handler concern (transport validation/mapping)? → handler/grpc/handler.go
   - Business rule / data validation? → service/transaction_service.go
   - SQL persistence? → repository/transaction_sql.go
   - Infrastructure (auth, metrics, retry)? → go-core, not here
   → If out of scope, explain where it belongs. Stop here.

2. AFFECTED LAYERS
   [ ] proto / generated code (requires `make proto`)
   [ ] handler (transport validation, proto ↔ domain mapping)
   [ ] service (business rules, error shaping)
   [ ] repository (SQL, dbtx.WithTx)
   [ ] migration (schema change → new numbered .sql file)
   [ ] domain (model or interface change)
   [ ] error_definitions (new error code → migration + service)

3. CONTRACT RISK
   [ ] Proto field added/removed → regenerate gen/go + consumer impact
   [ ] SQL schema change → idempotent migration required
   [ ] New error code → insert to transaction_error_definitions table
   [ ] StatusCode enum change → must stay in sync with SQL CHECK constraint
   [ ] reference_id uniqueness constraint touched → high blast radius

4. IMPLEMENTATION STEPS
   Ordered from lowest to highest risk.

5. FILES TO CHANGE
   List specific file paths and reason for each.

6. TESTS REQUIRED
   - Unit: service-level with sqlmock
   - Handler: request mapping + error path coverage

7. ACCEPTANCE CRITERIA
   - All existing tests pass
   - New behavior covered by tests
   - No raw SQL errors exposed to clients

Do not implement yet.
```
