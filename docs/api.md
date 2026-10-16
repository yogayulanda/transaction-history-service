# Transaction History API

Base URL: `http://localhost:8080`

All customer calls require authentication. The service derives the owner from the authenticated claims (`Subject` and `app_name` attribute); do not send owner identity in customer requests. HTTP responses use the shared envelope with `status`, `response_code`, `message`, `request_id`, `timestamp`, and `data`.

## List Transactions

`GET /v1/transactions`

| Query | Type | Meaning |
|---|---|---|
| `startDate` | `YYYY-MM-DD` | Inclusive start day in Asia/Jakarta. |
| `endDate` | `YYYY-MM-DD` | Inclusive end day in Asia/Jakarta. |
| `productCategory` | string | `transfer`, `ppob`, or `cico`. |
| `productType` | string | Product type filter. |
| `status` | string | Current status filter. |
| `cursor` | string | Opaque cursor returned by the previous page. |
| `pageSize` | integer | Defaults to 20; maximum 50. |

If both dates are omitted, the range is the latest 30 calendar days through today. The maximum range is 92 days. The end day is inclusive and is queried through the following Jakarta midnight. Invalid dates, reversed ranges, or invalid cursors return `CORE-1000`.

Example response:

```json
{
  "status": "SUCCESS",
  "response_code": "CORE-0000",
  "message": "success",
  "request_id": "req-sample-01",
  "timestamp": "2026-10-16T12:00:00Z",
  "data": {
    "items": [{
      "transactionId": "tx-sample-001",
      "productCategory": "transfer",
      "productType": "transfer_internal",
      "provider": "bca",
      "providerName": "BCA",
      "direction": "debit",
      "amount": "125000.00",
      "feeAmount": "2500.00",
      "totalAmount": "127500.00",
      "currency": "IDR",
      "status": "SUCCESS",
      "statusLabel": "Berhasil",
      "destinationNoMasked": "****4321",
      "transactionAt": "2026-10-16T04:30:00Z"
    }],
    "nextCursor": "eyJ0cmFuc2FjdGlvbl9hdCI6Ii4uLiIsImlkIjoidHgtc2FtcGxlLTAwMSJ9"
  }
}
```

Money values are decimal strings. Customer responses never include the partner, owner identity, raw destination, or source account number.

## Get Transaction

`GET /v1/transactions/{transaction_id}`

The transaction is returned only when it belongs to the authenticated owner. An unknown ID and another owner's ID return the same `CORE-1004` response. The detail includes masked account values, status history, and one typed detail object for the product type.

## Backoffice Search

`HistoryBackofficeService.SearchTransactions` is gRPC-only and is not exposed through HTTP. Claims must include role `HISTORY_BACKOFFICE_ROLE` (default `backoffice`). Search supports transaction ID, reference number, destination number plus type, or app plus user; at least one criterion is required. Date bounds use the same Jakarta calendar-day rules. Results are capped at 100 per page.

Destination values are normalized and HMAC-hashed before querying. Audit events record criterion names, result count, and at most the first eight hash characters; they never include the submitted destination number.

## Statuses and Errors

Statuses: `PENDING`, `UNKNOWN`, `SUCCESS`, `FAILED`, `NEED_REVIEW`, `REFUNDED`.

| Code | HTTP | Meaning |
|---|---:|---|
| `CORE-1000` | 400 | Invalid request. |
| `CORE-1001` | 401 | Authentication required. |
| `CORE-1003` | 403 | Caller is not authorized. |
| `CORE-1004` | 404 | Transaction not found for this owner. |
| `CORE-9000` | 500 | Internal service error. |

User-facing messages and status labels may be overridden by active message-provider rows; code defaults remain available.
