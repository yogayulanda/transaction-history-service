---
id: knowledge.errors-resilience
title: Errors and Resilience
type: knowledge
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: internal/service/errors.go }
  - { type: code, ref: internal/service/error_definition_resolver.go }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
  - { type: code, ref: internal/service/idempotency.go }
owner: unresolved
updated: 2026-07-06
---

# Errors and Resilience

## Service Error Taxonomy
- Validation invalid input: `TRH-VAL-001`
- Duplicate reference id: `TRH-VAL-002`
- Transaction history not found: `TRH-DB-001`
- Internal/recoverability bucket: `TRH-REC-001`

## Error Mapping Behavior
- Transport handlers convert service errors with `coreerrors.ToGRPC`.
- Duplicate `reference_id` from the repository maps to a non-retryable invalid-request error.
- Missing detail maps to a not-found error.
- Repository/runtime failures map to internal errors with sanitized user messages.

## Dynamic Error Definitions
- On startup, the app attempts to load active rows from `transaction_error_definitions`.
- Definitions refresh in background every `60s`.
- Resolver failures during bootstrap are logged but do not abort service startup.
- Resolver lookup can override user-facing message and field-level details for known TRH error codes.

## Kafka Retry and DLQ Policy
- Malformed events, unsupported event type/version, producer-not-allowed, invalid payload shape, invalid metadata JSON, invalid transaction time, and conflicting duplicate reference cases are treated as non-retryable.
- Non-retryable events are sent to the DLQ topic when a publisher is available.
- If DLQ publish fails, the handler returns a retryable error.
- Ordinary repository/runtime failures remain retryable and are not manually DLQ-published.
