INSERT INTO dbo.transaction_histories (
    id,
    user_id,
    reference_id,
    external_ref_id,
    product_group,
    product_type,
    transaction_route,
    channel,
    direction,
    amount,
    fee,
    total_amount,
    currency,
    status_code,
    error_code,
    error_message,
    source_service,
    transaction_time,
    created_at,
    updated_at
)
VALUES (
    'seed-tx-001',
    'seed-user-001',
    'seed-ref-001',
    'seed-ext-ref-001',
    'transfer',
    'transfer_internal',
    'internal',
    'mobile_app',
    'debit',
    10000,
    0,
    10000,
    'IDR',
    'SUCCESS',
    NULL,
    NULL,
    'seed',
    SYSUTCDATETIME(),
    SYSUTCDATETIME(),
    SYSUTCDATETIME()
);

INSERT INTO dbo.transaction_history_details (
    transaction_id,
    metadata_json,
    created_at,
    updated_at
)
VALUES (
    'seed-tx-001',
    '{"note":"seed data for local/manual testing"}',
    SYSUTCDATETIME(),
    SYSUTCDATETIME()
);

INSERT INTO dbo.transaction_history_status_events (
    transaction_id,
    from_status_code,
    to_status_code,
    reason_code,
    reason_message,
    event_time,
    raw_payload_json,
    created_at
)
VALUES (
    'seed-tx-001',
    'CREATED',
    'SUCCESS',
    NULL,
    NULL,
    SYSUTCDATETIME(),
    '{"source":"manual_seed"}',
    SYSUTCDATETIME()
);
