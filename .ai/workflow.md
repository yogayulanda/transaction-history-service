# Workflow

## Standard Change Loop

1. Identify scope and files.
2. Update code in one layer at a time.
3. Add/update tests close to changed behavior.
4. Run:
   - `go test ./...`
5. If dependencies changed:
   - `go mod tidy`
6. Update `.ai` context files if behavior changed.

## Scope Checklist

- API contract changed -> update proto/docs context
- Persistence changed -> update migration + transactions context
- Auth/config changed -> update security + integrations context
