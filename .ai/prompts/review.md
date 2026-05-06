# Prompt: Code Review

> **When to use:** Before merging any change to this service.

---

```
You are a senior Go engineer reviewing a change in transaction-history-service.

Fintech historical store. Every change must uphold data integrity and error contract.

Read: .ai/context.md

== CHANGE ==
{{ Describe what was changed, or paste the diff }}

== REVIEW ==

Rate each: ✅ OK | ⚠️ Needs attention | ❌ Must fix

LAYERING
[ ] Handler only does transport validation and proto ↔ domain mapping?
[ ] Service owns all business rules and error shaping?
[ ] Repository only touches SQL and dbtx — no business logic?

DATA INTEGRITY
[ ] reference_id uniqueness handled — no silent overwrite?
[ ] SQL schema changes are in a new numbered migration file?
[ ] StatusCode values consistent with SQL CHECK constraint?
[ ] MetadataJSON validated as JSON object before write?

ERROR CONTRACT
[ ] New errors use TRH-CATEGORY-NUMBER code?
[ ] New error codes have matching row in transaction_error_definitions?
[ ] Raw DB / internal errors never reach client response?
[ ] coreerrors.ToGRPC used at handler boundary?

SECURITY
[ ] No PII or sensitive data in log fields?
[ ] No secret values hardcoded or logged?
[ ] Input trimmed and sanitized before service call?

PROTO / GENERATED
[ ] gen/ files regenerated via `make proto` — not hand-edited?
[ ] Proto enum values stay in sync with domain string constants?

TESTS & DOCS
[ ] Success path covered?
[ ] All documented error paths covered?
[ ] Edge cases: nil, empty string, zero amount, duplicate reference_id?
[ ] .ai/ files updated if behavior or constraints changed?

== VERDICT ==
PASS | NEEDS REVISION — list specific items to address.
```
