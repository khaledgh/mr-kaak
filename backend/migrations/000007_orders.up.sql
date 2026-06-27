-- Orders (plan §4, §9). Money is integer cents; prices are snapshotted at order
-- time so later catalog edits never change historical orders.

CREATE TABLE orders (
    id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id              BIGINT UNSIGNED NOT NULL,
    code                 VARCHAR(20)     NOT NULL,
    status               VARCHAR(24)     NOT NULL DEFAULT 'pending_payment',
    fulfillment_type     ENUM('delivery','pickup') NOT NULL DEFAULT 'delivery',
    subtotal_cents       BIGINT          NOT NULL DEFAULT 0,
    discount_cents       BIGINT          NOT NULL DEFAULT 0,
    delivery_fee_cents   BIGINT          NOT NULL DEFAULT 0,
    tax_cents            BIGINT          NOT NULL DEFAULT 0,
    total_cents          BIGINT          NOT NULL DEFAULT 0,
    currency             VARCHAR(3)      NOT NULL DEFAULT 'CAD',
    payment_method       VARCHAR(20)     NOT NULL DEFAULT 'cod',
    payment_status       VARCHAR(24)     NOT NULL DEFAULT 'pending',
    coupon_id            BIGINT UNSIGNED NULL,
    address_snapshot_json JSON           NULL,
    scheduled_for        TIMESTAMP       NULL,
    notes                VARCHAR(500)    NULL,
    idempotency_key      VARCHAR(80)     NULL,
    created_at           TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at           TIMESTAMP       NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_orders_code (code),
    -- Idempotency: a (user, key) pair maps to exactly one order, so a retried
    -- checkout can never create a duplicate. NULL keys are allowed and not unique.
    UNIQUE KEY uq_orders_idem (user_id, idempotency_key),
    KEY idx_orders_user (user_id),
    KEY idx_orders_status (status),
    KEY idx_orders_deleted_at (deleted_at),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_orders_coupon FOREIGN KEY (coupon_id) REFERENCES coupons (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE order_items (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id         BIGINT UNSIGNED NOT NULL,
    product_id       BIGINT UNSIGNED NOT NULL,
    variant_id       BIGINT UNSIGNED NULL,
    name_snapshot    VARCHAR(200)    NOT NULL,
    qty              INT             NOT NULL DEFAULT 1,
    weight_grams     INT             NULL,
    unit_price_cents BIGINT          NOT NULL DEFAULT 0,
    modifiers_json   JSON            NULL,
    line_total_cents BIGINT          NOT NULL DEFAULT 0,
    created_at       TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_order_items_order (order_id),
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE order_status_history (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    order_id    BIGINT UNSIGNED NOT NULL,
    from_status VARCHAR(24)     NULL,
    to_status   VARCHAR(24)     NOT NULL,
    actor_id    BIGINT UNSIGNED NULL,
    note        VARCHAR(255)    NULL,
    created_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_status_history_order (order_id),
    CONSTRAINT fk_status_history_order FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
