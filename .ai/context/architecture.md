Architecture: Clean Architecture

Layers:

handler → service → repository

handler:

- receives HTTP or gRPC requests
- performs request validation
- maps requests to service calls

service:

- contains business logic
- orchestrates operations

repository:

- handles database access
- persists and retrieves data

Rules:

handler must not access database directly.

service must not depend on transport layer.

repository must not contain business logic.