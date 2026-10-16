---
id: knowledge.open-questions
title: Open Questions
type: knowledge
status: unknown
confidence: medium
source: hybrid
evidence:
  - { type: code, ref: .env.example }
  - { type: code, ref: cmd/server/main.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-10-16
---

# Open Questions

- `U-OWN` `[blocking]`: Repository owner and approval authority are not discoverable from current code or docs.
- `U-DEPLOY-001` `[blocking]`: Production deployment topology, environment mapping, and rollback procedure are not documented.
- `U-SEC-001` `[blocking]`: Confirm production internal-JWT issuer, audience, key source, and required gRPC metadata behavior; default claims resolver is blocked outside local without JWT.
- `U-OPS-001` `[informational]`: No on-call path, dashboard inventory, alert ownership, or SLO is documented.
- `U-DATA-001` `[important]`: Retention, deletion, and audit requirements for transaction and status history data are not documented.
- `U-PREDEPLOY-001` `[blocking]`: Apply migrations 0001-0005 to SQL Server, validate SQL Server lock-hint/unique-conflict behavior, verify rollback, and smoke both customer routes and the authenticated gRPC-only backoffice service.
- `U-PREDEPLOY-002` `[blocking]`: Run `make check` in the target release environment and review current govulncheck findings.
- `U-PREDEPLOY-003` `[blocking]`: Verify broker TLS/authentication, topic ACLs, DLQ permissions, configured source allowlist, and HMAC secret rotation procedure.
- `U-PREDEPLOY-004` `[blocking]`: Confirm gateway local endpoint transport settings and termination budget against deployment policy.
