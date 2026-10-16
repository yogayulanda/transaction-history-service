# Transaction History Service

This service stores transaction snapshots and ordered status history for customer history views and controlled support searches. It consumes status events; it does not process transactions or calculate money. Amounts remain decimal strings from producer to API.

## Runtime

- Go `1.26.9`; module directive `go 1.26.0`.
- Shared runtime: `github.com/yogayulanda/go-core` v`1.2.1`.
- SQL Server database: `transaction_history`; migrations run during startup.
- HTTP customer API: `GET /v1/transactions` and `GET /v1/transactions/{transaction_id}`.
- gRPC-only support API: `history.v1.HistoryBackofficeService/SearchTransactions`.
- Kafka inbound is controlled by `KAFKA_INBOUND_ENABLED`; its topic and producer allowlist are configured separately.

## Local Setup

Copy `.env.example` to `.env`, configure SQL Server, and set `HISTORY_HASH_KEY` to a hex-encoded secret containing at least 32 bytes. If the event consumer is enabled, also set `HISTORY_STATUS_TOPIC` and `HISTORY_ALLOWED_SOURCES`.

```bash
make run
```

Customer requests require authenticated claims. The service derives `user_id` from the claim subject and `app_name` from the `app_name` claim attribute; these values are not accepted as customer API parameters. Backoffice search requires the configured role, default `backoffice`.

The HTTP gateway forwards through the local gRPC server, so the same authentication, request-ID, logging, and error interceptors apply to HTTP and gRPC customer requests. Backoffice search is not registered on the HTTP gateway.

HTTP responses use the shared runtime envelope with `status`, `response_code`, `message`, `request_id`, `timestamp`, and `data`; validation responses may include `details`.

## Event Consumer

The event contract, product detail schemas, ordering rules, and DLQ behavior are documented in [docs/events/transaction-status-v1.md](docs/events/transaction-status-v1.md). Invalid events are sent to the topic's `.dlq` suffix with `x-dlq-reason`; database failures remain retryable. Raw destination/source numbers, CICO codes, and notes are not written to logs.

## Error Codes

| Code | HTTP | Meaning |
|---|---:|---|
| `CORE-1000` | 400 | Invalid request. |
| `CORE-1001` | 401 | Authentication is missing or invalid. |
| `CORE-1003` | 403 | The caller is not allowed. |
| `CORE-1004` | 404 | Transaction history was not found for this owner. |
| `CORE-9000` | 500 | An internal error occurred. |

Active `transaction_error_definitions` rows can override user-facing messages and status labels.

## Repository structure

```text
cmd/server/main.go
internal/
	app/                        # dependency construction and service registration
	domain/                     # types and pure rules; standard library only
		transaction.go
		status.go
		product.go
		event.go
		catalog.go
		repository.go
		errors.go
	service/                    # business flows; domain and utils dependencies
		status_event_service.go
		history_service.go
		backfill_service.go
		message_provider.go
		errors.go
	repository/                 # SQL Server access and static catalog
		transaction_repository.go
		message_repository.go
		catalog_static.go
	handler/
		grpc/                     # customer and backoffice RPCs, identity, mapping
		kafka/                    # CloudEvents/JSON decoding and DLQ
	utils/                      # pure standard-library helpers, no internal imports
		pii.go
		decimal.go
		timezone.go
```

Dependency direction: `handler → service → domain ← repository`; `app` wires dependencies, and `utils` imports nothing internal.

Where does new code go?
- SQL → `repository`
- Rule, list, or constant → `domain`
- Sequence of steps → `service`
- Request, JSON, or proto → `handler`
- Generic pure helper → `utils`
- Environment or wiring → `app`

## Developer Commands

```bash
make test
make vet
make lint
make vuln
make fmt-check
make check
```

Regenerate protobuf bindings with `make generate`. Required tools are `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, and `protoc-gen-grpc-gateway`.
