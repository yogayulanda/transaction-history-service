Task: harden transaction history service foundation

Goal:

Align repository foundation with the historical-store scope:
- clean runtime dependency usage
- production-safe migration baseline
- accurate developer workflow
- clear root documentation

Scope Layers:

bootstrap
contract
documentation

Allowed Paths:

README.md
.env.example
Makefile
cmd/server
internal/app
migrations/transaction
proto/history

Do NOT modify:

../go-core
external services

Constraints:

follow clean architecture
use go-core infrastructure
prefer minimal changes
keep create API as fallback/manual ingestion path

Expected Output:

- root README for the service
- synchronized local workflow and runtime wiring
- migration baseline without production seed coupling
- updated proto source comments aligned to product scope
