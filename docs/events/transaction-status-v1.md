# Transaction Status Event v1

This service consumes one complete transaction snapshot per status change. The Kafka key is `transaction_id`; producers must increment `ce_sequence` monotonically for each transaction. The service stores the highest accepted version and keeps the event history. The retired `transaction.created` flow is not consumed.

## CloudEvents Headers

| Header | Required | Type and rule |
|---|---|---|
| `ce_specversion` | Yes | `1.0`. |
| `ce_id` | Yes | UUID; unique event identity. |
| `ce_type` | Yes | Non-empty event type, at most 128 characters. |
| `ce_source` | Yes | CloudEvents source attribute. |
| `ce_time` | Yes | RFC3339 timestamp; persisted as `changed_at`. |
| `ce_subject` | Yes | Must equal payload `transaction_id`. |
| `ce_sequence` | Yes | Positive integer, strictly increasing per transaction. |

Attributes are read from the consumer context restored by the shared messaging runtime. Missing or invalid attributes are rejected; payload fields are not used as a fallback.

## Payload

Payload field names are snake_case. Unless marked optional, a field is required.

| Field | Type | Rule |
|---|---|---|
| `transaction_id` | string | Required; max 64 characters; equals `ce_subject`. |
| `app_name` | string | Required; max 16 characters; must be allowed by configuration. |
| `user_id` | string | Required; max 64 characters. |
| `product_category` | string | Required; max 16 characters; `transfer`, `ppob`, or `cico`; must match product type. |
| `product_type` | string | Required; max 32 characters; one of the types below. |
| `provider` | string | Required; max 32 characters; provider must be in the public catalog. |
| `partner` | string | Optional; max 32 characters; public sample catalog accepts `sample_partner`. |
| `direction` | string | Required; max 6 characters; `debit` or `credit`. |
| `denomination` | decimal string | Optional; same format as `amount`. |
| `amount`, `fee_amount`, `total_amount` | decimal string | Required; `^\d{1,16}(\.\d{1,2})?$`; SQL precision is `DECIMAL(18,2)`. |
| `currency` | string | Optional; defaults to `IDR`; three uppercase letters. |
| `status` | string | Required; max 16 characters; `PENDING`, `UNKNOWN`, `SUCCESS`, `FAILED`, `NEED_REVIEW`, or `REFUNDED`. |
| `from_status` | string | Optional; one of the status values. |
| `response_code` | string | Optional; max 16 characters. |
| `source_account_no` | string | Optional; digits only, 4–34 digits; only a masked value is stored. |
| `destination_name` | string | Optional; max 100 characters. |
| `destination_no` | string | Optional; requires `destination_no_type`; normalized, masked, and HMAC-hashed before persistence. |
| `destination_no_type` | string | Required with destination number; max 8 characters; `account`, `phone`, `meter`, or `va`. |
| `reference_no` | string | Optional; max 64 characters. |
| `source_system` | string | Required; max 32 characters; must be in `HISTORY_ALLOWED_SOURCES`. |
| `transaction_at` | string | Required; RFC3339 business timestamp. |
| `changed_by_type` | string | Required; max 16 characters; `user`, `system`, or `operator`. |
| `detail` | object | Optional; unknown keys and header-name keys are discarded. |

Phone normalization removes spaces, dashes, dots, and a leading plus; `62…` becomes `0…`. The normalized number must match `^08\d{7,13}$`. Account, meter, and VA values must contain 4–34 digits. HMAC-SHA256 input is `<destination_no_type>:<normalized value>` using the hex-decoded `HISTORY_HASH_KEY`.

## Product Details

All detail objects use `schema_version: 1` internally. Invalid types or values reject the event; unrecognized keys are dropped. Empty normalized details do not create a detail row.

| Product type | Category | Accepted detail keys |
|---|---|---|
| `transfer_bifast`, `transfer_rtol`, `transfer_internal` | `transfer` | `bank_code` (digits, max 8), `note` (max 100). |
| `pulsa` | `ppob` | `serial_number` (max 64). |
| `data_package` | `ppob` | `product_name` (max 100), `serial_number` (max 64). |
| `pln_prepaid` | `ppob` | `token` (max 32), `kwh` (max 16), `tariff_power` (max 32). |
| `pln_postpaid` | `ppob` | `period` (max 32), `stand_meter` (max 64). |
| `bpjs_kesehatan` | `ppob` | `period_months` (integer 1–24), `participant_count` (integer 1–50), `branch` (max 64). |
| `cash_in`, `cash_out` | `cico` | Input `code` is stored only as `code_masked` (last three characters visible); `warkat_no` (max 32). |

Fictional detail examples:

```json
{"product_type":"transfer_bifast","detail":{"bank_code":"014","note":"Monthly transfer"}}
{"product_type":"transfer_rtol","detail":{"bank_code":"008","note":"Invoice 204"}}
{"product_type":"transfer_internal","detail":{"bank_code":"002","note":"Savings transfer"}}
{"product_type":"pulsa","detail":{"serial_number":"SAMPLE-PULSA-001"}}
{"product_type":"data_package","detail":{"product_name":"Sample 10 GB","serial_number":"SAMPLE-DATA-001"}}
{"product_type":"pln_prepaid","detail":{"token":"12345678901234567890","kwh":"20.5","tariff_power":"R1-1300"}}
{"product_type":"pln_postpaid","detail":{"period":"2026-10","stand_meter":"001200-001350"}}
{"product_type":"bpjs_kesehatan","detail":{"period_months":1,"participant_count":2,"branch":"Sample City"}}
{"product_type":"cash_in","detail":{"code":"SAMPLE-CASH-IN-42","warkat_no":"WR-001"}}
{"product_type":"cash_out","detail":{"code":"SAMPLE-CASH-OUT-51","warkat_no":"WR-002"}}
```

## Ordering, Retry, and DLQ

- A repeated `ce_id` is acknowledged as a duplicate.
- A version lower than or equal to the current snapshot is recorded as stale without changing the header.
- A higher version replaces the snapshot. A final-to-non-final change is applied and logged as a regression anomaly.
- A unique conflict on `(transaction_id, status_version)` with a different event ID is logged as a version conflict and acknowledged.
- Validation failures are published to `<topic>.dlq` with `x-dlq-reason`, then acknowledged after a successful publish.
- Database failures remain retryable. A failed DLQ publish is also retryable and prevents acknowledgment.

Raw destination/source numbers, raw CICO codes, and `note` text are never logged. Logs contain stable reason codes and, when needed for search audit, only the first eight characters of the destination HMAC.