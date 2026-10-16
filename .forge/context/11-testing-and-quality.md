---
id: knowledge.testing-quality
title: Testing and Quality
type: knowledge
status: confirmed
confidence: high
source: hybrid
evidence:
  - { type: code, ref: Makefile }
  - { type: code, ref: internal/repository/history_sql_test.go }
  - { type: code, ref: internal/handler/grpc/history_handler_test.go }
  - { type: code, ref: internal/handler/kafka/status_event_test.go }
  - { type: code, ref: internal/backfill/backfill_test.go }
owner: unresolved
updated: 2026-10-16
---

# Testing and Quality

## Test Entry
- `make test` runs `go test -race ./...`.
- `make check` runs formatting, vet, race tests, golangci-lint, and govulncheck.

## Covered Areas
- PII normalization, masking, type-prefixed HMAC, catalog, and per-product detail validation.
- CloudEvents parsing, masked snapshot preparation, invalid-event DLQ, and retryable DB failures.
- SQL ordering cases: first event, higher version, stale event, duplicate event ID, and version conflict.
- Owner identity, missing claims, identical not-found behavior, Jakarta date bounds, cursors, status labels, and all detail oneofs.
- HTTP gateway authentication/interceptor path, request-ID forwarding, and backoffice route isolation.
- Backfill reruns, existing-row preservation, live-event updates, invalid counts, Jakarta grouping, and exact decimal totals.

## Remaining Validation Limits
- No live SQL Server/Kafka end-to-end harness is present.
- Migration rollback/live SQL semantics and production-like gateway/auth configuration require environment smoke tests.
