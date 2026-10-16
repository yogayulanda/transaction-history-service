---
id: knowledge.observability-support
title: Observability and Support
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/handler/grpc/backoffice_handler.go }
  - { type: code, ref: internal/service/status_event_service.go }
  - { type: code, ref: .env.example }
owner: unresolved
updated: 2026-10-16
---

# Observability and Support

## Logging
- Status-event service logs use stable operation/status/error codes and do not include snapshot payloads.
- Repository anomaly logs identify regression, duplicate event, and version conflict without raw PII.
- Every backoffice search emits `LogEvent{Category: "audit", Action: "transaction_search", Outcome: ...}` with criteria names and result count.
- Destination search audit may contain only the first eight characters of the HMAC hash.
- Startup describes available customer HTTP/gRPC routes and configured ports.

## Runtime Signals
- go-core provides health/readiness/version/metrics endpoints when configured.
- Startup readiness logging observes gRPC and HTTP ports.
- Tracing, metrics, diagnostics, and pprof are controlled by runtime configuration.

## Support Gaps
- No repository-local alert rules, dashboards, SLOs, or on-call escalation path are defined.
