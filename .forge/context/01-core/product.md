---
id: core.product
title: Product Context
type: core
status: inferred
confidence: high
source: ai
evidence:
  - { type: doc, ref: README.md }
  - { type: doc, ref: .ai/context.md }
owner: TBD
updated: 2026-05-20
---

# Product

## Summary

`transaction-history-service` adalah service penyimpan histori transaksi terpusat. Bertindak sebagai *historical store*, bukan transaction processor dan bukan reporting engine.

## Domain & Problem Space

- Domain: financial transaction history persistence.
- Problem: aplikasi dan downstream consumers butuh storage histori transaksi yang konsisten, queryable, dan dapat diandalkan sebagai fallback ingestion path saat jalur utama (event/Kafka) bermasalah.

## Users & Stakeholders

- Aplikasi (UI menu riwayat transaksi)
- Downstream reporting consumers
- Tim integrasi yang menggunakan service ini sebagai fallback/manual ingestion
- Tim QA pada tahap awal integrasi (testing insert)

## Producers (Data Sources)

- `trxFinance`
- `ms-liquiditas`
- `agent-payment-purchase`

## System Boundaries

### IN Scope

- Menyimpan transaksi via internal API `CreateTransactionHistory`
- Menyajikan daftar riwayat per user (`GetUserHistory`)
- Menyajikan detail satu transaksi (`GetTransactionHistoryDetail`)
- Menjadi historical store yang siap dikonsumsi app dan reporting downstream

### OUT of Scope

- Memproses transaksi finansial
- Menghitung report agregat bisnis di service ini
- Menggantikan jalur event/Kafka utama

## Core Product Terms

- **reference_id** — business transaction id, unik lintas semua producer
- **source_service** — producer yang mengirim transaksi
- **channel** — dimensi bisnis sumber transaksi (multi-app)
- **product_group**, **product_type** — taksonomi produk transaksi
- **transaction_route** — jalur transaksi
- **status_code** — status transaksi (enum sinkron dengan SQL constraints)
- **metadata_json** — JSON object untuk atribut spesifik produk; kolom inti TIDAK boleh dipindah ke sini
