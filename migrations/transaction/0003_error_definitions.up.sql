-- +goose Up
CREATE TABLE dbo.transaction_error_definitions (
    error_code VARCHAR(32) NOT NULL PRIMARY KEY,
    user_message NVARCHAR(512) NOT NULL,
    details_json NVARCHAR(MAX) NULL,
    is_active BIT NOT NULL CONSTRAINT DF_transaction_error_definitions_is_active DEFAULT (1),
    updated_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_error_definitions_updated_at DEFAULT (SYSUTCDATETIME())
);

-- +goose Down
DROP TABLE IF EXISTS dbo.transaction_error_definitions;
