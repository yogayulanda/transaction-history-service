# Security

## Runtime Guards (via go-core config)

- Internal JWT verification can be enabled/disabled by env.
- Signature validation can be enabled/disabled by env.
- HTTP pprof endpoint can be enabled/disabled by env.

## Service Rules

- Never log secrets or raw auth credentials.
- Keep error responses sanitized; use app error contract.
- Do not weaken auth checks in handlers.

## Config Surface

- `INTERNAL_JWT_*`
- `AUTH_SIGNATURE_*`
- `HTTP_PPROF_ENABLED`

Treat these as security-impacting settings.
