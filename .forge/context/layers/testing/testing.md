---
id: layer.testing.content
title: Testing Layer Conventions
type: layer
status: inferred
confidence: high
source: ai
evidence:
  - { type: code, ref: internal/handler/grpc/handler_test.go }
  - { type: code, ref: internal/service/transaction_service_test.go }
  - { type: code, ref: internal/repository/transaction_sql_test.go }
  - { type: code, ref: go.mod }
owner: unresolved
updated: 2026-05-20
---

# Testing Layer

## Test Strategy

Co-located unit tests per layer. No separate `tests/` directory.

| Layer | Test File | Pattern |
|---|---|---|
| Handler | `internal/handler/grpc/handler_test.go` | Mock service; verify request validation, error mapping |
| Service | `internal/service/transaction_service_test.go` | Mock repository; verify business rules, error shaping |
| Service | `internal/service/error_definition_resolver_test.go` | Pure logic test |
| Service | `internal/service/errors_test.go` | Error contract tests |
| Repository | `internal/repository/transaction_sql_test.go` | `go-sqlmock` for DB layer |

## Tooling

- Test runner: `go test ./...` (via `make test`)
- DB mocking: `github.com/DATA-DOG/go-sqlmock v1.5.2`

## Conventions

- Test files sibling to source: `<file>_test.go`.
- Table-driven tests preferred for multi-case logic.
- Mock at layer boundary, not within layer.
- Add tests with behavior changes (per `.ai/conventions.md`).

## Coverage Expectations

- Handler: input validation paths, error mapping paths.
- Service: business validation, error contract building, repository error handling.
- Repository: SQL query correctness, transactional boundary, mocked DB behavior.
- Domain: pure logic if any (entities, sentinels typically don't need tests).

## What is NOT Tested Here

- Integration tests against real SQL Server.
- End-to-end gRPC tests.
- Load/performance tests.

> Recorded in `knowledge/unknowns.md`: integration test strategy.
