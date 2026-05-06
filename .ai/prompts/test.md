# Prompt: Write Tests

> **When to use:** Adding or updating tests for a layer or specific behavior.

---

```
You are a senior Go engineer writing tests for transaction-history-service.

Fintech historical store. Tests guard data integrity and error contract correctness.
No real external services allowed — use sqlmock, interface mocks, httptest.

Read: .ai/context.md

== SCOPE ==
{{ Name the file / function to test, or "full coverage" for the whole layer }}

== LAYER RULES ==

Handler tests:
- Mock the service interface
- Cover: nil request, missing required fields, invalid enum, valid happy path
- Assert coreerrors.ToGRPC maps to correct gRPC status code

Service tests:
- Mock domain.TransactionRepository and domain.ErrorDefinitionRepository
- Cover: all required field validations, duplicate reference_id, not found, internal error
- Assert AppError code and HTTP status are correct for each case

Repository tests:
- Use go-sqlmock (DATA-DOG/go-sqlmock)
- Cover: insert success, duplicate key, not found, list with filters

== RULES ==
- Table-driven tests for all multi-case scenarios
- Test names: Test<FunctionName>_<Scenario>
- Test public behavior — not internal implementation
- No real DB, Redis, or external service calls
- Must pass: go test ./...

== OUTPUT ==
1. Complete test file — paste-ready, all imports included
2. Coverage summary — paths covered and edge cases included
3. Gaps — any behavior that cannot be tested without refactoring
```
