# transaction-history-service

`transaction-history-service` adalah service penyimpan histori transaksi untuk kebutuhan:

- menu riwayat transaksi di aplikasi
- downstream reporting
- fallback/manual ingestion ketika jalur utama ingest bermasalah
- testing insert data pada tahap awal integrasi

Service ini bukan transaction processor dan bukan reporting engine. Ia berfungsi sebagai historical store terpusat.

## Sumber Data

Producer transaksi yang menjadi sumber data utama:

- `trxFinance`
- `ms-liquiditas`
- `agent-payment-purchase`

## Scope Resmi

Yang termasuk scope:

- menyimpan transaksi baru melalui API internal `CreateTransactionHistory`
- mengambil daftar riwayat transaksi per user
- mengambil detail satu transaksi
- menyediakan data historis yang siap dipakai app dan consumer reporting lain

Yang tidak termasuk scope:

- memproses transaksi finansial
- menghitung report agregat bisnis di service ini
- menggantikan jalur event/Kafka utama

## Field Inti

Field inti yang wajib diperlakukan sebagai kolom utama, bukan dipindahkan ke `metadata_json`:

- `reference_id`
- `source_service`
- `channel`
- `product_group`
- `product_type`
- `transaction_route`
- `status_code`
- `transaction_time`

Catatan penting:

- `reference_id` adalah business transaction id dan harus unik lintas semua producer.
- `channel` adalah dimensi bisnis penting karena transaksi bisa datang dari lebih dari satu app.
- `metadata_json` tetap fleksibel, tetapi harus berupa JSON object yang valid dan hanya dipakai untuk atribut spesifik produk.

## Struktur Repo

```text
cmd/server                 entrypoint service
internal/app               wiring app service
internal/handler/grpc      transport layer gRPC/gateway
internal/service           business validation and orchestration
internal/repository        persistence layer
proto/history/v1           proto source
gen/go/history/v1          generated protobuf/gRPC/gateway code
migrations/transaction     database migration
.ai/                       AI workflow, context, task scope
```

## Setup Lokal

### 1. Siapkan environment

Gunakan `.env.example` sebagai baseline:

```bash
cp .env.example .env
```

Konfigurasi utama yang perlu diisi:

- SQL Server untuk database `transaction_history`
- `GRPC_PORT`
- `HTTP_PORT`
- `MIGRATION_AUTO_RUN` bila ingin migration jalan saat startup
- pengaturan auth/security sesuai kebutuhan (`INTERNAL_JWT_*`, `AUTH_SIGNATURE_*`, `HTTP_PPROF_ENABLED`)

### 2. Jalankan migration

Migration utama ada di `migrations/transaction`.

- `0001_init_transaction_history.up.sql` berisi baseline schema
- `0002_seed_transaction_history.up.sql` no-op agar production baseline bersih
- `migrations/dev/dev_seed_transaction_history.sql` dipakai hanya untuk local/manual seed

### 3. Generate protobuf

```bash
make proto
```

Kebutuhan tool lokal:

- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- `protoc-gen-grpc-gateway`

### 4. Jalankan test

```bash
make test
```

### 5. Jalankan service

```bash
make run
```

## API Ringkas

API utama service:

- `CreateTransactionHistory`
- `GetUserHistory`
- `GetTransactionHistoryDetail`

HTTP gateway dari `go-core` juga tersedia:

- `GET /health`
- `GET /ready`
- `GET /version`
- `GET /metrics`

## Catatan Kontrak Runtime

- `CreateTransactionHistory` dipertahankan sebagai fallback/manual ingestion path dan jalur testing insert.
- `GetUserHistory` menggunakan cursor offset placeholder (`nextCursor` berupa string angka offset).
- Saat `startDate` dan `endDate` diisi, handler memvalidasi `startDate <= endDate`.
- `GetUserHistory` mendukung filter status transaksi melalui field gRPC `status_code` dan query HTTP `statusCode`.
- Nilai status yang valid: `CREATED`, `PENDING`, `PROCESSING`, `SUCCESS`, `FAILED`, `REVERSED`, `EXPIRED`.
- `status_code` adalah nama field kontrak gRPC untuk transaction status filter; nilai `UNSPECIFIED` berarti tanpa filter status.
- Validasi field wajib business untuk create terjadi di service layer (termasuk `channel`, `sourceService`, `currency`, `statusCode`, dll).

## Error, Auth, dan Security

- Error service dibentuk sebagai `go-core/errors.AppError` dan dipetakan oleh handler via `coreerrors.ToGRPC`.
- Response error ke client disanitasi oleh kontrak error `go-core`.
- JWT enforcement dan signature middleware dikontrol oleh konfigurasi `go-core`, bukan middleware custom di service ini.

## Known Limitations

- pagination list masih placeholder
- lifecycle status update di luar create flow belum lengkap
- report agregat tetap menjadi tanggung jawab downstream consumer/job/service lain
- generated protobuf perlu diregenerate bila `proto/` berubah
