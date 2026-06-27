-- Users (accounts + RBAC) and their Canadian addresses.

CREATE TABLE users (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name               VARCHAR(120)    NOT NULL,
    email              VARCHAR(255)    NOT NULL,
    phone_e164         VARCHAR(20)     NULL,
    password_hash      VARCHAR(255)    NOT NULL,
    role               VARCHAR(20)     NOT NULL DEFAULT 'customer',
    status             VARCHAR(20)     NOT NULL DEFAULT 'active',
    token_version      INT             NOT NULL DEFAULT 0,
    default_address_id BIGINT UNSIGNED NULL,
    created_at         TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at         TIMESTAMP       NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email),
    KEY idx_users_phone (phone_e164),
    KEY idx_users_role (role),
    KEY idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE addresses (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id       BIGINT UNSIGNED NOT NULL,
    label         VARCHAR(60)     NULL,
    line1         VARCHAR(200)    NOT NULL,
    line2         VARCHAR(200)    NULL,
    city          VARCHAR(120)    NOT NULL,
    province_code VARCHAR(2)      NOT NULL,
    postal_code   VARCHAR(7)      NOT NULL,
    country_code  VARCHAR(2)      NOT NULL DEFAULT 'CA',
    lat           DECIMAL(10,7)   NULL,
    lng           DECIMAL(10,7)   NULL,
    phone_e164    VARCHAR(20)     NULL,
    notes         VARCHAR(255)    NULL,
    created_at    TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_addresses_user (user_id),
    KEY idx_addresses_deleted_at (deleted_at),
    CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Default address points back into addresses; added after both tables exist.
ALTER TABLE users
    ADD CONSTRAINT fk_users_default_address
    FOREIGN KEY (default_address_id) REFERENCES addresses (id) ON DELETE SET NULL;
