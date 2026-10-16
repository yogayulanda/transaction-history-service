-- +goose Up
-- +goose StatementBegin
IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'CORE-1001')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('CORE-1001', 'Sesi Anda telah berakhir. Silakan masuk kembali.', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'CORE-1003')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('CORE-1003', 'Anda tidak memiliki akses.', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'CORE-1004')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('CORE-1004', 'Riwayat transaksi tidak ditemukan.', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'CORE-9000')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('CORE-9000', 'Sistem sedang sibuk. Silakan coba beberapa saat lagi.', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'STATUS-PENDING')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('STATUS-PENDING', 'Sedang diproses', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'STATUS-UNKNOWN')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('STATUS-UNKNOWN', 'Sedang diproses', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'STATUS-SUCCESS')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('STATUS-SUCCESS', 'Berhasil', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'STATUS-FAILED')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('STATUS-FAILED', 'Gagal', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'STATUS-NEED_REVIEW')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('STATUS-NEED_REVIEW', 'Sedang diperiksa', NULL, 1, SYSUTCDATETIME());

IF NOT EXISTS (SELECT 1 FROM dbo.transaction_error_definitions WHERE error_code = 'STATUS-REFUNDED')
    INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
    VALUES ('STATUS-REFUNDED', 'Dana dikembalikan', NULL, 1, SYSUTCDATETIME());
-- +goose StatementEnd

-- +goose Down
DELETE FROM dbo.transaction_error_definitions
WHERE error_code IN (
    'CORE-1001', 'CORE-1003', 'CORE-1004', 'CORE-9000',
    'STATUS-PENDING', 'STATUS-UNKNOWN', 'STATUS-SUCCESS', 'STATUS-FAILED',
    'STATUS-NEED_REVIEW', 'STATUS-REFUNDED'
);