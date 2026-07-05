---
id: system.context-index
title: Context Index
type: core
status: confirmed
confidence: high
source: ai
evidence:
  - { type: code, ref: cmd/server/main.go }
  - { type: code, ref: internal/app/app.go }
  - { type: doc, ref: README.md }
owner: unresolved
updated: 2026-07-06
---

# Context Index

## Repository Profile
- Profile: service
- Context profile version: 2
- Primary language: Go

## How to Use This Context
- Start here, then load only the files that match the change surface.
- Treat `99-open-questions.md` as the stop list for anything still unknown.

## Read Paths

### Feature Implementation
Read:
- `01-service-overview.md`
- `02-architecture.md`
- `04-interfaces-and-contracts.md`
- `05-data-and-persistence.md`
- `06-business-rules-and-flows.md`
- `11-testing-and-quality.md`
- `99-open-questions.md`

### API or Integration Change
Read:
- `04-interfaces-and-contracts.md`
- `07-integrations-and-dependencies.md`
- `08-security-and-access.md`
- `09-errors-and-resilience.md`
- `12-runtime-deployment-and-config.md`
- `99-open-questions.md`

### Data Model Change
Read:
- `05-data-and-persistence.md`
- `06-business-rules-and-flows.md`
- `09-errors-and-resilience.md`
- `11-testing-and-quality.md`
- `99-open-questions.md`

### Kafka Ingestion Change
Read:
- `02-architecture.md`
- `06-business-rules-and-flows.md`
- `07-integrations-and-dependencies.md`
- `09-errors-and-resilience.md`
- `12-runtime-deployment-and-config.md`
- `99-open-questions.md`

### Incident or Runtime Investigation
Read:
- `09-errors-and-resilience.md`
- `10-observability-and-support.md`
- `12-runtime-deployment-and-config.md`
- `13-operations-and-runbook.md`
- `99-open-questions.md`

### Refactor
Read:
- `02-architecture.md`
- `03-domain-boundaries.md`
- `11-testing-and-quality.md`
- `14-decisions-assumptions-and-constraints.md`
- `99-open-questions.md`
