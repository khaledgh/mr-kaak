-- Web Push subscriptions (plan §8). One row per browser push endpoint.
CREATE TABLE push_subscriptions (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL,
    endpoint   VARCHAR(500)    NOT NULL,
    p256dh     VARCHAR(255)    NOT NULL,
    auth       VARCHAR(255)    NOT NULL,
    user_agent VARCHAR(255)    NULL,
    created_at TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_push_endpoint (endpoint),
    KEY idx_push_user (user_id),
    CONSTRAINT fk_push_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
