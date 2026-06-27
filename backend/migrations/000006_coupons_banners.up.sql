-- Coupons (plan §11) and marketing banners (plan §12).

CREATE TABLE coupons (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code             VARCHAR(60)     NOT NULL,
    type             ENUM('percent','fixed','free_delivery') NOT NULL,
    value            BIGINT          NOT NULL DEFAULT 0,  -- percent: 0-100; fixed: cents
    min_order_cents  BIGINT          NOT NULL DEFAULT 0,
    max_discount_cents BIGINT        NOT NULL DEFAULT 0,  -- 0 = uncapped (percent cap)
    usage_limit      INT             NOT NULL DEFAULT 0,  -- 0 = unlimited (global)
    used_count       INT             NOT NULL DEFAULT 0,
    per_user_limit   INT             NOT NULL DEFAULT 0,  -- 0 = unlimited (per user)
    starts_at        TIMESTAMP       NULL,
    ends_at          TIMESTAMP       NULL,
    is_active        TINYINT(1)      NOT NULL DEFAULT 1,
    created_at       TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMP       NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_coupons_code (code),
    KEY idx_coupons_active (is_active),
    KEY idx_coupons_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- One row per successful redemption; enforces per_user_limit and gives analytics.
CREATE TABLE coupon_redemptions (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    coupon_id  BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    order_id   BIGINT UNSIGNED NULL,
    created_at TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_redemptions_coupon_user (coupon_id, user_id),
    CONSTRAINT fk_redemptions_coupon FOREIGN KEY (coupon_id) REFERENCES coupons (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE banners (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title      VARCHAR(160)    NOT NULL,
    image_url  VARCHAR(500)    NOT NULL,
    link_url   VARCHAR(500)    NULL,
    sort_order INT             NOT NULL DEFAULT 0,
    starts_at  TIMESTAMP       NULL,
    ends_at    TIMESTAMP       NULL,
    is_active  TINYINT(1)      NOT NULL DEFAULT 1,
    created_at TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_banners_active_sort (is_active, sort_order),
    KEY idx_banners_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
