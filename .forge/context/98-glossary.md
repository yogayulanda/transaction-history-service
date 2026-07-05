---
id: knowledge.glossary
title: Glossary
type: knowledge
status: confirmed
confidence: high
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: code, ref: proto/history/v1/history.proto }
  - { type: code, ref: internal/handler/kafka/transaction_created.go }
owner: unresolved
updated: 2026-07-06
---

# Glossary

- `transaction-history-service`: Service-owned historical store for transaction records.
- `reference_id`: Business transaction identifier expected to be unique across producer systems.
- `source_service`: Upstream producer/service name persisted with each transaction.
- `metadata_json`: Product-specific JSON object payload stored in `transaction_history_details`.
- `HistoryService`: The gRPC service defined in `proto/history/v1/history.proto`.
- `DLQ`: Dead-letter topic `transaction-history.transaction.created.dlq` for non-retryable Kafka inbound failures.
- `TRH`: Error taxonomy domain prefix used by this service (`TRH-VAL-*`, `TRH-DB-*`, `TRH-REC-*`).
