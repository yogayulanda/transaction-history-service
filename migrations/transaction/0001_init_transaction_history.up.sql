-- +goose Up
CREATE TABLE dbo.transaction_histories (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    app_name VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    reference_no VARCHAR(128) NULL,
    product_category VARCHAR(16) NOT NULL,
    product_type VARCHAR(64) NOT NULL,
    provider VARCHAR(64) NOT NULL,
    partner VARCHAR(64) NULL,
    direction VARCHAR(16) NOT NULL,
    denomination VARCHAR(64) NULL,
    amount DECIMAL(18,2) NOT NULL,
    fee_amount DECIMAL(18,2) NOT NULL,
    total_amount DECIMAL(18,2) NOT NULL,
    currency CHAR(3) NOT NULL CONSTRAINT DF_transaction_histories_currency DEFAULT ('IDR'),
    status VARCHAR(32) NOT NULL,
    status_version BIGINT NOT NULL,
    response_code VARCHAR(64) NULL,
    source_account_masked VARCHAR(64) NULL,
    destination_name NVARCHAR(128) NULL,
    destination_no_type VARCHAR(16) NULL,
    destination_no_masked VARCHAR(64) NULL,
    destination_no_hash CHAR(64) NULL,
    source_system VARCHAR(64) NOT NULL,
    transaction_at DATETIME2 NOT NULL,
    completed_at DATETIME2 NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_histories_created_at DEFAULT (SYSUTCDATETIME()),
    updated_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_histories_updated_at DEFAULT (SYSUTCDATETIME()),

    CONSTRAINT CK_transaction_histories_product_category
        CHECK (product_category IN ('transfer', 'ppob', 'cico')),
    CONSTRAINT CK_transaction_histories_direction
        CHECK (direction IN ('debit', 'credit')),
    CONSTRAINT CK_transaction_histories_status
        CHECK (status IN ('PENDING', 'UNKNOWN', 'SUCCESS', 'FAILED', 'NEED_REVIEW', 'REFUNDED')),
    CONSTRAINT CK_transaction_histories_destination_no_type
        CHECK (destination_no_type IS NULL OR destination_no_type IN ('account', 'phone', 'meter', 'va')),
    CONSTRAINT CK_transaction_histories_amount_non_negative
        CHECK (amount >= 0 AND fee_amount >= 0 AND total_amount >= 0)
);

CREATE INDEX IX_th_user_list
    ON dbo.transaction_histories(app_name, user_id, transaction_at DESC, id DESC)
    INCLUDE (product_type, provider, denomination, direction, total_amount, status, destination_name, destination_no_masked);

CREATE TABLE dbo.transaction_history_details (
    transaction_id VARCHAR(64) NOT NULL PRIMARY KEY,
    schema_version INT NOT NULL,
    fields_json NVARCHAR(MAX) NOT NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_details_created_at DEFAULT (SYSUTCDATETIME()),
    updated_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_details_updated_at DEFAULT (SYSUTCDATETIME()),

    CONSTRAINT FK_transaction_history_details_transaction
        FOREIGN KEY (transaction_id) REFERENCES dbo.transaction_histories(id) ON DELETE CASCADE
);


CREATE TABLE dbo.transaction_history_status_events (
    id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
    transaction_history_id VARCHAR(64) NOT NULL,
    event_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    status_version BIGINT NOT NULL,
    from_status VARCHAR(32) NULL,
    to_status VARCHAR(32) NOT NULL,
    response_code VARCHAR(64) NULL,
    changed_by_type VARCHAR(16) NOT NULL,
    changed_at DATETIME2 NOT NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_status_events_created_at DEFAULT (SYSUTCDATETIME()),

    CONSTRAINT FK_transaction_history_status_events_transaction
        FOREIGN KEY (transaction_history_id) REFERENCES dbo.transaction_histories(id) ON DELETE CASCADE,
    CONSTRAINT CK_transaction_history_status_events_to_status
        CHECK (to_status IN ('PENDING', 'UNKNOWN', 'SUCCESS', 'FAILED', 'NEED_REVIEW', 'REFUNDED')),
    CONSTRAINT CK_transaction_history_status_events_from_status
        CHECK (from_status IS NULL OR from_status IN ('PENDING', 'UNKNOWN', 'SUCCESS', 'FAILED', 'NEED_REVIEW', 'REFUNDED')),
    CONSTRAINT CK_transaction_history_status_events_changed_by_type
        CHECK (changed_by_type IN ('user', 'system', 'operator'))
);

CREATE UNIQUE INDEX UX_transaction_history_status_events_event_id
    ON dbo.transaction_history_status_events(event_id);

CREATE UNIQUE INDEX UX_transaction_history_status_events_version
    ON dbo.transaction_history_status_events(transaction_history_id, status_version);

CREATE INDEX IX_transaction_history_status_events_tx_time
    ON dbo.transaction_history_status_events(transaction_history_id, changed_at DESC, id DESC);

-- +goose Down
IF OBJECT_ID('dbo.transaction_history_status_events', 'U') IS NOT NULL
BEGIN
    DROP TABLE dbo.transaction_history_status_events;
END;

IF OBJECT_ID('dbo.transaction_history_details', 'U') IS NOT NULL
BEGIN
    DROP TABLE dbo.transaction_history_details;
END;

IF OBJECT_ID('dbo.transaction_histories', 'U') IS NOT NULL
BEGIN
    DROP TABLE dbo.transaction_histories;
END;
