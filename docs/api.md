# Transaction History API

This service exposes both gRPC and HTTP/REST interfaces. The REST endpoints are generated via gRPC Gateway, and the endpoints provided below represent the JSON REST API mapping.

## Base URL
`http://localhost:8080/v1`

---

## 1. Get User History
Retrieves a paginated list of transaction history records for a given user.

**Endpoint:** `GET /history`

### Query Parameters

| Field | Type | Required | Description | Example |
|---|---|---|---|---|
| `userId` | `string` | No | End user identifier. | `usr_123` |
| `startDate` | `string` | No | Inclusive start date (ISO8601). | `2026-02-01T00:00:00Z` |
| `endDate` | `string` | No | Inclusive end date (ISO8601). | `2026-02-29T23:59:59Z` |
| `productGroup` | `string` | No | Top-level product grouping for reporting. | `transfer` |
| `productType` | `string` | No | Specific product/use-case under `productGroup`. | `transfer_internal` |
| `transactionRoute`| `string` | No | Backend route used to process transaction. | `internal` |
| `statusCode` | `string` | No | Current lifecycle status code. | `TRANSACTION_STATUS_CODE_SUCCESS` |
| `cursor` | `string` | No | Pagination cursor from previous response. | |
| `pageSize` | `integer`| No | Requested page size (max 100). | `10` |

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
      "sourceService": "transfer_service",
      "transactionTime": "2026-04-10T12:00:00Z"
    }
  ],
  "nextCursor": "eyJwYWdlIjoxfQ==",
  "hasMore": false
}
```

---

## 2. Get Transaction History Detail
Returns the detailed information of a specific transaction history record.

**Endpoint:** `GET /history/{id}`

### Path Parameters

| Field | Type | Required | Description | Example |
|---|---|---|---|---|
| `id` | `string` | Yes | Internal immutable transaction ID (UUID/ULID). | `123e4567-e89b-12d3-a456-426614174000` |

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
    "sourceService": "transfer_service",
    "transactionTime": "2026-04-10T12:00:00Z",
    "metadataJson": "{\"beneficiary_account\": \"1234567890\"}",
    "createdAt": "2026-04-10T12:00:00.123Z",
    "updatedAt": "2026-04-10T12:00:00.123Z"
  }
}
```

---

## 3. Create Transaction History
Internal and fallback API endpoint to directly ingest a transaction history record into the persistent store.

**Endpoint:** `POST /transaction-histories`

### Request Body

| Field | Type | Required | Description |
|---|---|---|---|
| `userId` | `string` | Yes* | End user identifier. |
| `referenceId` | `string` | Yes* | Idempotency/business reference ID from caller system. Must be globally unique. |
| `externalRefId` | `string` | No | Optional provider/core-banking reference ID. |
| `productGroup` | `string` | Yes* | Top-level product grouping (e.g. `transfer`). |
| `productType` | `string` | Yes* | Product/use-case (e.g. `transfer_internal`). |
| `transactionRoute` | `string` | No | Processing route in backend. |
| `channel` | `string` | Yes* | Origin channel of request (e.g. `mobile_app`). |
| `direction` | `string` | Yes* | Financial direction from account perspective: `debit`, `credit`. |
| `amount` | `string(int64)`| Yes* | Monetary values stored in minor unit (e.g. cents). |
| `fee` | `string(int64)`| No | Transaction fee. |
| `totalAmount` | `string(int64)`| Yes* | Amount + Fee. |
| `currency` | `string` | Yes* | ISO currency code (e.g., `IDR`). |
| `statusCode` | `string` | Yes* | Lifecycle status code. |
| `errorCode` | `string` | No | Optional business/integration error code. |
| `errorMessage` | `string` | No | Optional business/integration error description. |
| `sourceService` | `string` | No | Upstream source service name. |
| `transactionTime`| `string(date-time)`| Yes* | Business transaction time (ISO8601). |
| `metadataJson` | `string` | No | Product-specific payload in escaped JSON format. |

_(**\*** Marks standard required fields for business usage. Protobuf 3 standard marks everything optional loosely, so required implies logically required rules expected from the downstream service layers)._

### Response Example

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

## Reference: Status Codes
| Value | Description |
|---|---|
| `TRANSACTION_STATUS_CODE_UNSPECIFIED` | Default fallback. |
| `TRANSACTION_STATUS_CODE_CREATED` | Just created, before processing. |
| `TRANSACTION_STATUS_CODE_PENDING` | Request sent to downstream, awaiting callback. |
| `TRANSACTION_STATUS_CODE_PROCESSING` | Actively processing the request internally. |
| `TRANSACTION_STATUS_CODE_SUCCESS` | Transaction completed successfully. |
| `TRANSACTION_STATUS_CODE_FAILED` | Transaction failed synchronously or asynchronously. |
| `TRANSACTION_STATUS_CODE_REVERSED` | Processed transaction that was rolled back/voided. |
| `TRANSACTION_STATUS_CODE_EXPIRED` | Payment or sequence expired prior to completion. |
