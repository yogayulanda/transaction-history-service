-- +goose Up
CREATE TABLE dbo.transaction_histories (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    reference_id VARCHAR(128) NOT NULL,
    external_ref_id VARCHAR(128) NULL,
    product_group VARCHAR(32) NOT NULL,
    product_type VARCHAR(64) NOT NULL,
    transaction_route VARCHAR(32) NOT NULL,
    channel VARCHAR(32) NOT NULL,
    direction VARCHAR(16) NOT NULL,
    amount BIGINT NOT NULL,
    fee BIGINT NOT NULL CONSTRAINT DF_transaction_histories_fee DEFAULT (0),
    total_amount BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    status_code VARCHAR(32) NOT NULL,
    error_code VARCHAR(64) NULL,
    error_message NVARCHAR(512) NULL,
    source_service VARCHAR(64) NOT NULL CONSTRAINT DF_transaction_histories_source_service DEFAULT (''),
    transaction_time DATETIME2 NOT NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_histories_created_at DEFAULT (SYSUTCDATETIME()),
    updated_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_histories_updated_at DEFAULT (SYSUTCDATETIME()),

    CONSTRAINT CK_transaction_histories_product_group
        CHECK (product_group IN ('ppob', 'transfer', 'cash')),
    CONSTRAINT CK_transaction_histories_transaction_route
        CHECK (transaction_route IN ('internal', 'bifast', 'rtol', 'switching', 'partner_h2h')),
    CONSTRAINT CK_transaction_histories_direction
        CHECK (direction IN ('debit', 'credit')),
    CONSTRAINT CK_transaction_histories_status_code
        CHECK (status_code IN ('CREATED', 'PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'REVERSED', 'EXPIRED')),
    CONSTRAINT CK_transaction_histories_amount_non_negative
        CHECK (amount >= 0 AND fee >= 0 AND total_amount >= 0)
);

CREATE UNIQUE INDEX UX_transaction_histories_source_reference
    ON dbo.transaction_histories(source_service, reference_id);

CREATE INDEX IX_transaction_histories_user_time
    ON dbo.transaction_histories(user_id, transaction_time DESC);

CREATE INDEX IX_transaction_histories_status_time
    ON dbo.transaction_histories(status_code, transaction_time DESC);

CREATE INDEX IX_transaction_histories_product_type_time
    ON dbo.transaction_histories(product_type, transaction_time DESC);


CREATE TABLE dbo.transaction_history_details (
    transaction_id VARCHAR(64) NOT NULL PRIMARY KEY,
    metadata_json NVARCHAR(MAX) NOT NULL CONSTRAINT DF_transaction_history_details_metadata_json DEFAULT ('{}'),
    created_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_details_created_at DEFAULT (SYSUTCDATETIME()),
    updated_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_details_updated_at DEFAULT (SYSUTCDATETIME()),

    CONSTRAINT FK_transaction_history_details_transaction
        FOREIGN KEY (transaction_id) REFERENCES dbo.transaction_histories(id) ON DELETE CASCADE
);


CREATE TABLE dbo.transaction_history_status_events (
    id BIGINT IDENTITY(1,1) NOT NULL PRIMARY KEY,
    transaction_id VARCHAR(64) NOT NULL,
    from_status_code VARCHAR(32) NULL,
    to_status_code VARCHAR(32) NOT NULL,
    reason_code VARCHAR(64) NULL,
    reason_message NVARCHAR(512) NULL,
    event_time DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_status_events_event_time DEFAULT (SYSUTCDATETIME()),
    raw_payload_json NVARCHAR(MAX) NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT DF_transaction_history_status_events_created_at DEFAULT (SYSUTCDATETIME()),

    CONSTRAINT FK_transaction_history_status_events_transaction
        FOREIGN KEY (transaction_id) REFERENCES dbo.transaction_histories(id) ON DELETE CASCADE,
    CONSTRAINT CK_transaction_history_status_events_to_status_code
        CHECK (to_status_code IN ('CREATED', 'PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'REVERSED', 'EXPIRED')),
    CONSTRAINT CK_transaction_history_status_events_from_status_code
        CHECK (from_status_code IS NULL OR from_status_code IN ('CREATED', 'PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'REVERSED', 'EXPIRED'))
);

CREATE INDEX IX_transaction_history_status_events_tx_time
    ON dbo.transaction_history_status_events(transaction_id, event_time DESC);

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
