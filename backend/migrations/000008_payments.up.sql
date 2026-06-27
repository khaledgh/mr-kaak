-- Payment transactions (plan §10). One row per provider interaction (charge,
-- webhook event, refund), kept as the source of truth alongside the order.
CREATE TABLE payment_transactions (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id     BIGINT UNSIGNED NOT NULL,
    provider     VARCHAR(20)     NOT NULL,            -- 'square' | 'cod'
    provider_ref VARCHAR(120)    NULL,                -- provider payment/refund id
    kind         VARCHAR(20)     NOT NULL DEFAULT 'charge', -- charge | refund
    amount_cents BIGINT          NOT NULL DEFAULT 0,
    currency     VARCHAR(3)      NOT NULL DEFAULT 'CAD',
    status       VARCHAR(24)     NOT NULL,            -- pending|succeeded|failed|refunded
    raw_payload_json JSON        NULL,
    created_at   TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_payment_txn_order (order_id),
    -- Webhook idempotency: a provider_ref is processed at most once.
    UNIQUE KEY uq_payment_provider_ref (provider, provider_ref),
    CONSTRAINT fk_payment_txn_order FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed Square credential placeholders so the admin settings screen shows them.
INSERT INTO settings (`key`, value_json, description) VALUES
    ('square_environment',           '"sandbox"', 'Square environment: sandbox|production'),
    ('square_application_id',        '""',        'Square application id'),
    ('square_location_id',           '""',        'Square location id'),
    ('square_access_token',          '""',        'Square access token (secret)'),
    ('square_webhook_signature_key', '""',        'Square webhook signature key (secret)')
ON DUPLICATE KEY UPDATE description = VALUES(description);
