-- Admin-managed languages (plan §7). Adding a row here makes a new locale
-- available app-wide with no redeploy. is_default marks the fallback locale;
-- is_rtl drives RTL layout (Arabic).
CREATE TABLE languages (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code        VARCHAR(10)     NOT NULL,         -- 'en','ar','fr'
    name        VARCHAR(80)     NOT NULL,         -- English name
    native_name VARCHAR(80)     NOT NULL,         -- endonym
    is_default  TINYINT(1)      NOT NULL DEFAULT 0,
    is_rtl      TINYINT(1)      NOT NULL DEFAULT 0,
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    sort_order  INT             NOT NULL DEFAULT 0,
    created_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_languages_code (code),
    KEY idx_languages_active_sort (is_active, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO languages (code, name, native_name, is_default, is_rtl, is_active, sort_order) VALUES
    ('en', 'English', 'English', 1, 0, 1, 1),
    ('ar', 'Arabic',  'العربية', 0, 1, 1, 2),
    ('fr', 'French',  'Français', 0, 0, 1, 3);

-- UI string bundles per locale (static frontend strings), editable in admin.
CREATE TABLE ui_strings (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    locale      VARCHAR(10)     NOT NULL,
    `key`       VARCHAR(160)    NOT NULL,
    value       TEXT            NOT NULL,
    created_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_ui_string (locale, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
