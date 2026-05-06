# Transaction History API

Service exposes gRPC and HTTP/REST via grpc-gateway.

## Base URL

`http://localhost:8080/v1`

---

## 1. Get User History

Retrieve paginated transaction history for a user.

**Endpoint:** `GET /history`

### Query Parameters

| Field | Type | Required | Description | Example |
|---|---|---|---|---|
| `userId` | `string` | Yes | End user identifier. | `usr_123` |
| `startDate` | `string` | No | Inclusive start date (RFC3339). | `2026-02-01T00:00:00Z` |
| `endDate` | `string` | No | Inclusive end date (RFC3339). | `2026-02-29T23:59:59Z` |
| `productGroup` | `string` | No | Product group filter. | `transfer` |
| `productType` | `string` | No | Product type filter. | `transfer_internal` |
| `transactionRoute` | `string` | No | Processing route filter. | `internal` |
| `statusCode` | `string` | No | Lifecycle status code enum. | `TRANSACTION_STATUS_CODE_SUCCESS` |
| `cursor` | `string` | No | Offset cursor as non-negative integer string. | `20` |
| `pageSize` | `integer` | No | Requested page size (default 20, max 100). | `10` |

Validation notes:
- `startDate` and `endDate` must be valid RFC3339.
- If both are provided, `startDate` must be before or equal to `endDate`.

### Response Example

```json
{
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "userId": "usr_123",
      "referenceId": "ref_abc123",
      "externalRefId": "ext_ref_456",
      "productGroup": "transfer",
      "productType": "transfer_internal",
      "transactionRoute": "internal",
      "channel": "mobile_app",
      "direction": "debit",
      "amount": "100000",
      "fee": "0",
      "totalAmount": "100000",
      "currency": "IDR",
      "statusCode": "TRANSACTION_STATUS_CODE_SUCCESS",
      "errorCode": "",
      "errorMessage": "",
      "sourceService": "trxFinance",
      "transactionTime": "2026-04-10T12:00:00Z"
    }
  ],
  "nextCursor": "10",
  "hasMore": true
}
```

---

## 2. Get Transaction History Detail

Retrieve detail by transaction ID.

**Endpoint:** `GET /history/{id}`

### Path Parameters

| Field | Type | Required | Description | Example |
|---|---|---|---|---|
| `id` | `string` | Yes | Immutable transaction ID (UUID/ULID). | `123e4567-e89b-12d3-a456-426614174000` |

### Response Example

```json
{
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "userId": "usr_123",
    "referenceId": "ref_abc123",
    "externalRefId": "ext_ref_456",
    "productGroup": "transfer",
    "productType": "transfer_internal",
    "transactionRoute": "internal",
    "channel": "mobile_app",
    "direction": "debit",
    "amount": "100000",
    "fee": "0",
    "totalAmount": "100000",
    "currency": "IDR",
    "statusCode": "TRANSACTION_STATUS_CODE_SUCCESS",
    "errorCode": "",
    "errorMessage": "",
    "sourceService": "trxFinance",
    "transactionTime": "2026-04-10T12:00:00Z",
    "metadataJson": "{\"beneficiary_account\":\"1234567890\"}",
    "createdAt": "2026-04-10T12:00:00.123Z",
    "updatedAt": "2026-04-10T12:00:00.123Z"
  }
}
```

---

## 3. Create Transaction History

Internal fallback/manual ingestion endpoint.

**Endpoint:** `POST /transaction-histories`

### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `userId` | `string` | Yes | End user identifier. |
| `referenceId` | `string` | Yes | Business reference ID; must be unique. |
| `externalRefId` | `string` | No | Optional provider/core-banking reference ID. |
| `productGroup` | `string` | Yes | Top-level product group. |
| `productType` | `string` | Yes | Product/use-case type. |
| `transactionRoute` | `string` | No | Processing route. |
| `channel` | `string` | Yes | Origin channel. |
| `direction` | `string` | Yes | Direction (`debit` or `credit`). |
| `amount` | `string(int64)` | Yes | Amount in minor unit. |
| `fee` | `string(int64)` | No | Fee in minor unit. |
| `totalAmount` | `string(int64)` | Yes | Total amount in minor unit. |
| `currency` | `string` | Yes | ISO currency code (3 letters). |
| `statusCode` | `string` | Yes | Lifecycle status code enum. |
| `errorCode` | `string` | No | Optional integration error code. |
| `errorMessage` | `string` | No | Optional integration error message. |
| `sourceService` | `string` | Yes | Upstream source service name. |
| `transactionTime` | `string(date-time)` | No | Business transaction time; defaults to server UTC when omitted. |
| `metadataJson` | `string` | No | Optional JSON object payload as string. |

Validation notes:
- `metadataJson` must be a valid JSON object when provided.
- `amount`, `fee`, and `totalAmount` must be non-negative.

### Response Example

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

## Error and Security Notes

- Service error responses are sanitized through `go-core/errors` mapping (`coreerrors.ToGRPC`).
- gRPC and HTTP status semantics remain canonical (`InvalidArgument`, `NotFound`, `Internal`, etc.).
- JWT verification mode and signature middleware are controlled by go-core runtime config.

## Reference: Status Codes

| Value | Description |
|---|---|
| `TRANSACTION_STATUS_CODE_UNSPECIFIED` | Default fallback. |
| `TRANSACTION_STATUS_CODE_CREATED` | Just created, before processing. |
| `TRANSACTION_STATUS_CODE_PENDING` | Waiting downstream callback/processing. |
| `TRANSACTION_STATUS_CODE_PROCESSING` | Actively processing. |
| `TRANSACTION_STATUS_CODE_SUCCESS` | Completed successfully. |
| `TRANSACTION_STATUS_CODE_FAILED` | Failed synchronously/asynchronously. |
| `TRANSACTION_STATUS_CODE_REVERSED` | Processed then reversed/voided. |
| `TRANSACTION_STATUS_CODE_EXPIRED` | Expired before completion. |
