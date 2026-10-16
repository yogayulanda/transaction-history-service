---
id: system.security-access
title: Security and Access
type: system
system_type: service
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: internal/identity/identity.go }
  - { type: code, ref: internal/handler/grpc/backoffice_handler.go }
  - { type: code, ref: cmd/server/main.go }
  - { type: doc, ref: docs/events/transaction-status-v1.md }
owner: unresolved
updated: 2026-10-16
---

# Security and Access

## Identity and Authorization
- Customer owner resolves from go-core claims: `Subject` is user ID and `Attributes["app_name"]` is app name.
- Missing subject/app claims return `CORE-1001`; app names must pass the configured catalog.
- Backoffice search requires role `HISTORY_BACKOFFICE_ROLE` (default `backoffice`) or returns `CORE-1003`.
- Outside local environment, the default claims resolver requires internal JWT to be enabled; a fork may install a validated custom resolver.
- HTTP gateway forwards through local gRPC, so gRPC auth and request interceptors apply.

## Sensitive Data
- Destination account, phone, meter, and VA values are normalized, masked, and HMAC-hashed before persistence/search.
- Source account values are masked; CICO codes are stored only masked; detail keys matching header columns are dropped.
- Transfer `note` and PLN prepaid `token` are persisted unmasked in filtered detail fields by design and may be returned only to the transaction owner; a PLN token is usable only with that owner's meter.
- Raw account/phone/meter/VA numbers, CICO codes, PLN tokens, and notes must never appear in logs, error messages, or audit events.
- Invalid-event DLQ payloads remain unchanged for replay but are never logged. The internal fork must apply the source topic's ACL, TLS, and retention controls to its `.dlq` topic (I-A).
- `HISTORY_HASH_KEY` is hex-encoded secret material and must decode to at least 32 bytes.

## Producer Boundary
- Producer allowlist is configuration, not authentication; Kafka trust and broker access remain deployment responsibilities.
