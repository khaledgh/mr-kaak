-- Foundational key/value settings store. Holds runtime-configurable values
-- (payment toggles, tax %, store hours, Meta keys, languages config, ...)
-- that admins edit without a redeploy. value_json keeps each value typed.
CREATE TABLE settings (
    `key`        VARCHAR(128) NOT NULL,
    value_json   JSON         NOT NULL,
    description  VARCHAR(255) NULL,
    updated_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed the toggles the checkout/payment layer reads (filled in later phases).
INSERT INTO settings (`key`, value_json, description) VALUES
    ('cod_enabled',    'true',   'Cash on Delivery payment method enabled'),
    ('square_enabled', 'false',  'Square card payment method enabled'),
    ('tax_percent',    '0',      'Default sales tax percent (overridable per province)'),
    ('currency',       '"CAD"',  'Store currency code');
