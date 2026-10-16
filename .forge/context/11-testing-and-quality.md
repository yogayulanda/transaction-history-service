---
id: knowledge.testing-quality
title: Testing and Quality
type: knowledge
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: Makefile }
  - { type: code, ref: internal/service/transaction_service_test.go }
  - { type: code, ref: internal/service/idempotency_test.go }
  - { type: code, ref: internal/handler/grpc/handler_test.go }
  - { type: code, ref: internal/handler/kafka/transaction_created_test.go }
  - { type: code, ref: internal/repository/transaction_sql_test.go }
owner: unresolved
updated: 2026-10-15
---

# Testing and Quality

## Test Entry
- `make test` runs `go test -race ./...`.
- `make check` runs `fmt-check`, `vet`, race tests, `lint`, and `vuln`.
- `make lint` uses golangci-lint v2.14.0; `make vuln` runs `govulncheck ./...`.

## Covered Areas
- Service validation and normalization rules.
- Registered TRH response-code and gRPC mapping, CORE-1000 field details, and database-backed message-provider overrides.
- Idempotent duplicate handling semantics.
- gRPC handler validation and status-filter mapping.
- Kafka inbound classification, allowlist, DLQ behavior, and retryability.
- Repository SQL behavior using `sqlmock`, including duplicate key handling, list pagination, and error-definition reads.
- Kafka inbound config defaults.
- Observability identity field-bag propagation for create operations.

## Current Test Style
- Mostly unit tests with stubs/fakes and SQL mocks.
- No repository-local end-to-end test harness or deployment smoke tests were found.

## Quality Implication
- Business validation, transport mapping, and persistence query behavior are reasonably covered.
- Cross-service integration behavior with real `go-core`, real Kafka, or a live SQL Server instance is not covered in this repo.
- Migration idempotence and live HTTP envelope behavior remain pre-deploy smoke checks.
