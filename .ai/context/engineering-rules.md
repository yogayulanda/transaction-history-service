Engineering Rules

Always follow clean architecture.

handler → service → repository

Never access database directly from handler.

Service layer owns business logic.

Repository layer owns persistence logic.

Always use go-core infrastructure.

Never create custom bootstrap logic.

Prefer minimal change over large refactors.

Follow existing naming conventions.