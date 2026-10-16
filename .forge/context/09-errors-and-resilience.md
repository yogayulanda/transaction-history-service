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
updated: 2026-10-15
---

# Errors and Resilience

## Service Error Taxonomy
- Field validation: `CORE-1000` (`errors.Validation`); field violations are exposed as `details`.
- Duplicate reference id: `TRH-VAL-002` (HTTP 400, gRPC `InvalidArgument`).
- Transaction history not found: `TRH-DB-001` (HTTP 404, gRPC `NotFound`).
- Internal service failure: `TRH-REC-001` (HTTP 500, gRPC `Internal`).
- Unknown/unregistered errors fall back to `CORE-9000`.
- `TRH-VAL-001` is retired and unused; its existing database row is retained.

## Error Mapping Behavior
- Service codes are registered with `go-core/errors.Register`; transport handlers map them through `coreerrors.ToGRPC`.
- Duplicate `reference_id` from the repository maps to `TRH-VAL-002`.
- Missing detail maps to a not-found error.
- Repository/runtime failures map to `TRH-REC-001` with sanitized messages.
- Active `transaction_error_definitions.user_message` values implement the go-core message provider. `details_json` remains in the schema but is no longer applied to responses.

## Dynamic Error Definitions
- On startup, the app loads active rows from `transaction_error_definitions` as a go-core message provider.
- Definitions refresh in background every `60s`.
- Resolver failures during bootstrap are logged but do not abort service startup.
- Provider lookup can override registered default messages when a non-empty `user_message` exists. Field-level details come from `CORE-1000` validation, not database definitions.

## Kafka Retry and DLQ Policy
- Malformed events, unsupported event type/version, producer-not-allowed, invalid payload shape, invalid metadata JSON, invalid transaction time, conflicting duplicates, and registered non-retryable `ClassFailed` app errors are non-retryable.
- The service manually sends non-retryable events to the DLQ topic with `x-original-topic` and `x-error-code` headers.
- If DLQ publish fails, the handler returns a retryable error.
- Ordinary repository/runtime failures remain retryable and are not manually DLQ-published.
- When retries are exhausted, the go-core v1.2.1 consumer may publish to the DLQ with `x-dlq-reason` and `x-original-topic`; successful DLQ publication commits the original message, while failed publication leaves it uncommitted and blocks that partition.
- Processing and commits preserve order per Kafka partition; concurrency applies across partitions.
