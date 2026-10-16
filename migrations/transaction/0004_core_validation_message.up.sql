-- +goose Up
-- +goose StatementBegin
IF NOT EXISTS (
    SELECT 1
    FROM dbo.transaction_error_definitions
    WHERE error_code = 'CORE-1000'
)
BEGIN
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('CORE-1000', 'Permintaan tidak valid, silakan periksa kembali data yang dimasukkan.', NULL, 1, SYSUTCDATETIME());
END;
-- +goose StatementEnd

-- +goose Down
DELETE FROM dbo.transaction_error_definitions
WHERE error_code = 'CORE-1000';