# Prompt: Security Review

> **When to use:** Any change touching input handling, auth, error messages, logging, or SQL.

---

```
You are a security engineer reviewing a change in transaction-history-service.

This is a fintech service storing financial transaction records.
Data integrity and confidentiality are non-negotiable.

Read: .ai/context.md

== CHANGE ==
{{ Describe what was changed, or paste the diff }}

== AUDIT ==

Rate each: ✅ SAFE | ⚠️ NEEDS ATTENTION | ❌ VULNERABILITY

INPUT HANDLING
[ ] All string inputs trimmed before use?
[ ] Required fields validated before reaching repository?
[ ] MetadataJSON validated as JSON object — not arbitrary string?
[ ] Numeric fields (amount, fee, total_amount) validated non-negative?
[ ] Currency validated as 3-letter alpha code?
[ ] Date range: startDate <= endDate enforced before query?
[ ] Pagination bounds enforced (pageSize 1–100, offset >= 0)?

DATA PROTECTION
[ ] No PII (user_id, amount, reference_id) in log fields at INFO level?
[ ] Internal DB error detail does not leak into gRPC response messages?
[ ] error_message from DB row is user-facing — not an internal trace?
[ ] MetadataJSON content not logged in full?

AUTH & TRANSPORT
[ ] JWT enforcement delegated to go-core middleware — not re-implemented here?
[ ] HMAC signature validation delegated to go-core — not re-implemented here?
[ ] No auth bypass in handler for any RPC method?

SQL SAFETY
[ ] No raw string concatenation in SQL queries?
[ ] GORM parameterized queries used throughout?
[ ] Duplicate reference_id returns correct error — not a 500?

SECRETS
[ ] No secrets, DSN, or tokens hardcoded?
[ ] .env values never logged?

== VERDICT ==
SEVERITY: Critical | High | Medium | Low | None
APPROVED TO MERGE: YES / NO
Fixes required: (list ❌ and ⚠️ items with specific remediation)
```
