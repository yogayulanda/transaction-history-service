-- +goose Up
CREATE TABLE dbo.transaction_histories (
    id                      BIGINT IDENTITY(1,1) NOT NULL,
    transaction_id          VARCHAR(64)    NOT NULL,
    app_name                VARCHAR(16)    NOT NULL,
    user_id                 VARCHAR(64)    NOT NULL,
    product_category        VARCHAR(16)    NOT NULL,
    product_type            VARCHAR(32)    NOT NULL,
    provider                VARCHAR(32)    NOT NULL,
    partner                 VARCHAR(32)    NULL,
    direction               VARCHAR(6)     NOT NULL,
    denomination            DECIMAL(18,2)  NULL,
    amount                  DECIMAL(18,2)  NOT NULL,
    fee_amount              DECIMAL(18,2)  NOT NULL CONSTRAINT DF_th_fee_amount DEFAULT (0),
    total_amount            DECIMAL(18,2)  NOT NULL,
    currency                CHAR(3)        NOT NULL CONSTRAINT DF_th_currency DEFAULT ('IDR'),
    status                  VARCHAR(16)    NOT NULL,
    status_version          BIGINT         NOT NULL,
    response_code           VARCHAR(16)    NULL,
    source_account_masked   VARCHAR(32)    NULL,
    destination_name        NVARCHAR(100)  NULL,
    destination_no_type     VARCHAR(8)     NULL,
    destination_no_masked   VARCHAR(32)    NULL,
    destination_no_hash     CHAR(64)       NULL,
    reference_no            VARCHAR(64)    NULL,
    source_system           VARCHAR(32)    NOT NULL,
    transaction_at          DATETIME2(3)   NOT NULL,
    completed_at            DATETIME2(3)   NULL,
    created_at              DATETIME2(3)   NOT NULL CONSTRAINT DF_th_created_at DEFAULT (SYSUTCDATETIME()),
    updated_at              DATETIME2(3)   NOT NULL CONSTRAINT DF_th_updated_at DEFAULT (SYSUTCDATETIME()),
    CONSTRAINT PK_transaction_histories PRIMARY KEY CLUSTERED (id),
    CONSTRAINT CK_th_direction CHECK (direction IN ('debit','credit')),
    CONSTRAINT CK_th_status CHECK (status IN ('PENDING','UNKNOWN','SUCCESS','FAILED','NEED_REVIEW','REFUNDED')),
    CONSTRAINT CK_th_destination_no_type CHECK (destination_no_type IS NULL OR destination_no_type IN ('account','phone','meter','va')),
    CONSTRAINT CK_th_amounts CHECK (amount >= 0 AND fee_amount >= 0 AND total_amount >= 0 AND (denomination IS NULL OR denomination >= 0))
);

CREATE UNIQUE INDEX UX_th_transaction_id ON dbo.transaction_histories (transaction_id);

CREATE INDEX IX_th_user_list ON dbo.transaction_histories (app_name, user_id, transaction_at DESC, id DESC)
    INCLUDE (product_type, provider, denomination, direction, total_amount, status, destination_name, destination_no_masked);

CREATE INDEX IX_th_reference_no ON dbo.transaction_histories (reference_no)
    WHERE reference_no IS NOT NULL;

CREATE INDEX IX_th_destination_no_hash ON dbo.transaction_histories (destination_no_hash, transaction_at DESC)
    WHERE destination_no_hash IS NOT NULL;

CREATE INDEX IX_th_open ON dbo.transaction_histories (status, transaction_at)
    WHERE status IN ('PENDING','UNKNOWN','NEED_REVIEW');

CREATE TABLE dbo.transaction_history_details (
    transaction_history_id  BIGINT         NOT NULL,
    schema_version          SMALLINT       NOT NULL,
    detail_json             NVARCHAR(4000) NOT NULL,
    CONSTRAINT PK_transaction_history_details PRIMARY KEY CLUSTERED (transaction_history_id),
    CONSTRAINT FK_thd_transaction_history FOREIGN KEY (transaction_history_id) REFERENCES dbo.transaction_histories (id),
    CONSTRAINT CK_thd_detail_json CHECK (ISJSON(detail_json) = 1)
);

CREATE TABLE dbo.transaction_history_status_events (
    id                      BIGINT IDENTITY(1,1) NOT NULL,
    transaction_history_id  BIGINT         NOT NULL,
    event_id                VARCHAR(36)    NOT NULL,
    event_type              VARCHAR(128)   NOT NULL,
    status_version          BIGINT         NOT NULL,
    from_status             VARCHAR(16)    NULL,
    to_status               VARCHAR(16)    NOT NULL,
    response_code           VARCHAR(16)    NULL,
    changed_by_type         VARCHAR(16)    NOT NULL,
    changed_at              DATETIME2(3)   NOT NULL,
    received_at             DATETIME2(3)   NOT NULL CONSTRAINT DF_thse_received_at DEFAULT (SYSUTCDATETIME()),
    CONSTRAINT PK_transaction_history_status_events PRIMARY KEY CLUSTERED (id),
    CONSTRAINT FK_thse_transaction_history FOREIGN KEY (transaction_history_id) REFERENCES dbo.transaction_histories (id),
    CONSTRAINT CK_thse_to_status CHECK (to_status IN ('PENDING','UNKNOWN','SUCCESS','FAILED','NEED_REVIEW','REFUNDED')),
    CONSTRAINT CK_thse_from_status CHECK (from_status IS NULL OR from_status IN ('PENDING','UNKNOWN','SUCCESS','FAILED','NEED_REVIEW','REFUNDED')),
    CONSTRAINT CK_thse_changed_by_type CHECK (changed_by_type IN ('user','system','operator'))
);

CREATE UNIQUE INDEX UX_thse_event_id ON dbo.transaction_history_status_events (event_id);

CREATE UNIQUE INDEX UX_thse_status_version ON dbo.transaction_history_status_events (transaction_history_id, status_version);

CREATE INDEX IX_thse_history_time ON dbo.transaction_history_status_events (transaction_history_id, changed_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS dbo.transaction_history_status_events;
DROP TABLE IF EXISTS dbo.transaction_history_details;
DROP TABLE IF EXISTS dbo.transaction_histories;
