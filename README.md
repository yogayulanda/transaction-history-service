# Transaction History Service

`transaction-history-service` stores transaction history for application history views, downstream reporting, and controlled fallback/manual ingestion. It is a historical store, not a transaction processor or reporting engine.

Primary producers include `trxFinance`, `ms-liquiditas`, and the configured agent-payment producer. Core fields such as `reference_id`, `source_service`, `channel`, `product_group`, `product_type`, `transaction_route`, `status_code`, and `transaction_time` remain first-class columns. `metadata_json` is reserved for product-specific JSON objects.

## Runtime

- Go toolchain: `1.26.9`; module directive: `go 1.26.0`.
- Shared runtime: public `github.com/yogayulanda/go-core` module at `v1.2.1`; no sibling checkout is required.
- SQL Server database: `transaction_history`.
- Migration auto-run is configured in `cmd/server/main.go`; migrations run from `migrations/transaction` during startup.
- Tracing endpoint: `APM_ENDPOINT`. Metrics and diagnostics are configured with `METRICS_ENABLED` and `DIAG_ENABLED`.
- Kafka is enabled by a non-empty `KAFKA_BROKERS`; `KAFKA_INBOUND_ENABLED` controls this service's inbound consumer.
- A non-empty `AUTH_SIGNATURE_MASTER_KEY` enables signature validation. Internal JWT remains disabled for this service.

## Local Setup

Use `.env.example` as the baseline and configure SQL Server connection values for `transaction_history`. Migration settings are code-configured, not environment variables.

```bash
cp .env.example .env
make run
```

The service exposes gRPC and HTTP gateway APIs. The HTTP routes are:

- `GET /v1/history`
- `GET /v1/history/{id}`
- `POST /v1/transaction-histories` for internal fallback/manual ingestion

The go-core gateway also provides `GET /health`, `/ready`, `/version`, and `/metrics`.

HTTP JSON responses use the go-core v1.2.1 envelope: `status`, `response_code`, `message`, `request_id`, `timestamp`, `data`, and (for `CORE-1000`) `details`. The examples in [docs/api.md](docs/api.md), [docs/openapi.yaml](docs/openapi.yaml), and the Postman collection are derived from go-core v1.2.1 code and are pending smoke verification against a running service and SQL Server.

## Error Model

Service response codes are registered through `go-core/errors`. Active rows in `transaction_error_definitions` provide message overrides through the go-core message-provider interface; `details_json` remains in the database schema but is not applied. Field validation uses `CORE-1000` and carries field-level `details`.

| Code | HTTP | Meaning |
|---|---:|---|
| `CORE-1000` | 400 | Invalid request with optional field details. |
| `TRH-VAL-002` | 400 | Duplicate transaction reference. |
| `TRH-DB-001` | 404 | Transaction history not found. |
| `TRH-REC-001` | 500 | Non-recoverable service error. |
| `CORE-9000` | 500 | Fallback for unknown errors. |

## Developer Commands

```bash
make test       # race-enabled test suite
make vet
make lint       # golangci-lint v2.14.0
make vuln       # govulncheck
make fmt-check
make check      # formatting, vet, tests, lint, and vulnerability scan
```

Regenerate protobuf output with `make proto` only when the proto source or generator versions change. Required tools are `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, and `protoc-gen-grpc-gateway`.

## Contract Notes

- `GetUserHistory` uses an offset-based cursor placeholder and validates the date range when both dates are provided.
- The HTTP `statusCode` query maps to the gRPC `status_code` field; `UNSPECIFIED` means no status filter.
- Valid transaction statuses are `CREATED`, `PENDING`, `PROCESSING`, `SUCCESS`, `FAILED`, `REVERSED`, and `EXPIRED`.
- Kafka non-retryable events are dead-lettered; ordinary repository/runtime failures remain retryable.
- Lifecycle status updates and aggregate reporting are outside this service's current scope.
