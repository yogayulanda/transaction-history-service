-- +goose Up
CREATE TABLE dbo.transaction_error_definitions (
    error_code VARCHAR(32) NOT NULL PRIMARY KEY,
    user_message NVARCHAR(512) NOT NULL,
    details_json NVARCHAR(MAX) NULL,
    is_active BIT NOT NULL CONSTRAINT DF_transaction_error_definitions_is_active DEFAULT (1),
    updated_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_error_definitions_updated_at DEFAULT (SYSUTCDATETIME())
);

INSERT INTO dbo.transaction_error_definitions (error_code, user_message, details_json, is_active, updated_at)
VALUES
    ('TRH-VAL-001', 'Permintaan tidak valid, silakan periksa kembali data yang dimasukkan.', NULL, 1, SYSUTCDATETIME()),
    ('TRH-VAL-002', 'Transaksi dengan referensi ini sudah pernah diproses.', NULL, 1, SYSUTCDATETIME()),
    ('TRH-DB-001', 'Riwayat transaksi tidak ditemukan.', NULL, 1, SYSUTCDATETIME()),
    ('TRH-REC-001', 'Sistem sedang sibuk. Silakan coba beberapa saat lagi.', NULL, 1, SYSUTCDATETIME());

-- +goose Down
IF OBJECT_ID('dbo.transaction_error_definitions', 'U') IS NOT NULL
BEGIN
    DROP TABLE dbo.transaction_error_definitions;
END;
