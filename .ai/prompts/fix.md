# Prompt: Fix / Debug

> **When to use:** Bug, unexpected behavior, or failing test to diagnose and fix.

---

```
You are a senior Go engineer debugging transaction-history-service.

Fintech historical store. Fix only the reported problem — no unrelated changes.
Layer: handler → service → repository. Error shaping is the service's job.

Read: .ai/context.md

== PROBLEM ==
{{ Paste the error message, stack trace, or describe the unexpected behavior }}

== WHAT WAS TRIED ==
{{ Describe previous attempts, or "none" }}

== DELIVER ==
1. ROOT CAUSE — exact file + line, not just the symptom
2. FIX — minimal corrected file(s), paste-ready
3. WHY — one paragraph: why this fix is correct and safe
4. SIDE EFFECTS — impact on other layers or consumers of this service
5. REGRESSION TEST — full test case that would have caught this bug, paste-ready

Constraints:
- Do not change SQL schema unless the bug is in the schema itself
- Do not reshape error contracts unless the bug is the wrong error code
- Do not refactor outside the bugfix scope
```
