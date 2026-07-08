---
id: knowledge.open-questions
title: Open Questions
type: knowledge
status: unknown
confidence: medium
source: ai
evidence:
  - { type: code, ref: .env.example }
  - { type: code, ref: cmd/server/main.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-07-06
---

# Open Questions

- `U-OWN` `[blocking]`: Repository owner and approval authority are not discoverable from current code or docs.
- `U-DEPLOY-001` `[blocking]`: Production deployment topology, environment mapping, and rollback procedure are not documented in this repo.
- `U-SEC-001` `[blocking]`: The authoritative JWT/signature policy lives in `go-core` configuration behavior, but this repo does not state which auth modes are required in each environment.
- `U-OPS-001` `[informational]`: No on-call escalation path, dashboard inventory, or alert ownership is documented here.
- `U-DOMAIN-001` `[informational]`: README names producer systems, while Kafka code uses a slightly different producer identifier for agent-payment (`ms-agent-payment-purchase` vs `agent-payment-purchase` wording in docs); maintainers should confirm the canonical producer name set.
- `U-FLOW-001` `[informational]`: The schema includes status-event history and the domain defines `ErrInvalidStatus`, but no status-update flow beyond initial create is implemented in current service code.
- `U-DATA-001` `[informational]`: Compliance, retention, and audit requirements for transaction history data are not documented in this repo, even though the service stores user identifiers, transaction records, and append-only status events.
- `U-PERF-001` `[informational]`: No SLA, latency target, throughput expectation, or pagination-scale target is documented for `CreateTransactionHistory` or `GetUserHistory`.
- `U-VALIDATION-001` `[informational]`: Schema and docs require `direction`, but current gRPC/service/Kafka required-field validation does not clearly enforce non-empty `direction` before persistence; maintainers should confirm whether service-layer validation is intended to reject missing direction consistently.
