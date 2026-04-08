Task: implement transaction history endpoint

Goal:

Provide API endpoint to fetch transaction history by user_id.

Scope Layers:

repository
service
handler

Allowed Paths:

internal/repository
internal/service
internal/handler

Constraints:

follow clean architecture
use go-core database access

Expected Result:

repository query
service method
handler endpoint