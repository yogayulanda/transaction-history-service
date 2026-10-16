---
id: knowledge.decisions-assumptions-constraints
title: Decisions Assumptions and Constraints
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: doc, ref: .forge/generated/ecp/2026-10-16-th-v1-ecp.md }
  - { type: code, ref: internal/handler/kafka/status_event.go }
  - { type: code, ref: internal/identity/identity.go }
  - { type: code, ref: migrations/transaction/0001_init_transaction_history.up.sql }
owner: unresolved
updated: 2026-10-16
---

# Decisions Assumptions and Constraints

## Confirmed Decisions
- Money stays decimal strings through event, domain, and API; SQL uses `DECIMAL(18,2)` and service does no money arithmetic.
- Public identity resolution uses claims: subject to user ID and `app_name` attribute to app name. Missing identity maps to `CORE-1001`.
- HTTP gateway calls local gRPC so customer HTTP requests pass through gRPC interceptors.
- Backoffice is a separate gRPC service, never registered on the gateway, and requires the configured role.
- Core failure codes are `CORE-1000`, `CORE-1001`, `CORE-1003`, `CORE-1004`, and `CORE-9000`; no service-owned TRH codes remain.
- Status labels use `STATUS-<STATUS>` message rows with code fallbacks.
- Producer allowlist, topic, and HMAC key are configuration; public samples are fictional.
- Destination HMAC is type-prefixed; details are schema-filtered. Transfer notes and PLN prepaid tokens are intentionally persisted unmasked and shown only to the transaction owner; they are never logged.
- Invalid-event DLQ payloads are intentionally unchanged for replay and never logged. Internal fork I-A must apply the source topic's ACL, TLS, and retention controls to the `.dlq` topic.
- Raw destination/source numbers and CICO codes are not persisted; raw account/phone/meter/VA values, CICO codes, PLN tokens, and notes are never logged.
- Status event application is transactionally ordered by sequence; backfill is insert-only at version zero.

## Constraints
- Go `1.26.9`, directive `1.26.0`, go-core `v1.2.1`; no new direct dependency was needed.
- Generated protobuf output must be produced by `make generate`, not hand-edited.
- Do not stage, commit, push, or publish without explicit request.

## Assumptions
- go-core remains owner of shared auth, transport, lifecycle, and message-provider behavior.
- Production topic/source values and deployment topology are supplied outside this public repository.
