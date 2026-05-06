# Integrations and Dependencies

## Internal Framework

- `github.com/yogayulanda/go-core` (local replace: `../go-core`)
- Provides config, server, migration, logging, DB abstraction, auth/signature middleware, and error mapper.

## go-core Compatibility (Latest)

- Handler maps service errors with `coreerrors.ToGRPC`.
- Auth mode is framework-controlled:
  - JWT verification mode (`INTERNAL_JWT_ENABLED=true`)
  - metadata extraction mode (`INTERNAL_JWT_ENABLED=false`)
- Signature middleware is framework-controlled (`AUTH_SIGNATURE_ENABLED`).
- HTTP pprof exposure is framework-controlled (`HTTP_PPROF_ENABLED`).
- Service observability uses framework log flavors:
  - `ServiceLog`
  - `DBLog`
  - `TransactionLog`

## Data and Messaging

- SQL Server (required): `transaction_history`
- Redis cache (optional)
- Kafka publisher (optional when enabled)

## API and Tooling

- gRPC + grpc-gateway
- Proto/Buf generated code in `gen/go/...`
- Docs assets in `docs/` (`openapi.yaml`, `postman_collection.json`)

## Operational Notes

- `MIGRATION_AUTO_RUN=true` auto-applies migrations on startup.
- Keep `go.mod` and `go.sum` synchronized (`go mod tidy`).
