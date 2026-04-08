Service: transaction-history-service

Purpose:
Service responsible for storing and retrieving transaction history.

Core features:

- store transaction records
- retrieve transaction history by user_id
- expose APIs via gRPC and HTTP gateway

Repository structure:

cmd/
internal/
  handler/
  service/
  repository/
  model/

migrations/
proto/

Dependencies:

go-core
database
optional messaging

Notes:

This service runs on top of go-core framework.

All infrastructure must come from go-core.