INSERT INTO dbo.transaction_histories (
    id,
    app_name,
    user_id,
    reference_no,
    product_category,
    product_type,
    provider,
    partner,
    direction,
    amount,
    fee_amount,
    total_amount,
    currency,
    status,
    status_version,
    response_code,
    source_account_masked,
    destination_name,
    destination_no_type,
    destination_no_masked,
    destination_no_hash,
    source_system,
    transaction_at,
    completed_at
)
VALUES (
    'seed-tx-001',
    'app_a',
    'seed-user-001',
    'seed-ref-001',
    'transfer',
    'transfer_internal',
    'bca',
    NULL,
    'debit',
    NULL,
    10000.00,
    0.00,
    10000.00,
    'IDR',
    'SUCCESS',
    1,
    NULL,
    NULL,
    '****7890',
    'Sample recipient',
    'account',
    '****1234',
    REPLICATE('a', 64),
    'sample-producer',
    SYSUTCDATETIME(),
    SYSUTCDATETIME()
);

INSERT INTO dbo.transaction_history_details (
    transaction_id,
    schema_version,
    fields_json
)
VALUES (
    'seed-tx-001',
    1,
    '{"note":"sample transfer"}'
);

INSERT INTO dbo.transaction_history_status_events (
    transaction_history_id,
    event_id,
    event_type,
    status_version,
    from_status,
    to_status,
    response_code,
    changed_by_type,
    changed_at
)
VALUES (
    'seed-tx-001',
    '6ba7b810-9dad-11d1-80b4-00c04fd430c8',
    'transaction.status.v1',
    1,
    NULL,
    'SUCCESS',
    NULL,
    'system',
    SYSUTCDATETIME()
);
