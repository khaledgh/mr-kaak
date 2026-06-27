-- Catalog: categories, products (per-unit & per-kg), size variants, modifier
-- groups/modifiers, and a generic translations store for i18n names/descriptions.

CREATE TABLE categories (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    slug        VARCHAR(140)    NOT NULL,
    sort_order  INT             NOT NULL DEFAULT 0,
    image_url   VARCHAR(500)    NULL,
    is_active   TINYINT(1)      NOT NULL DEFAULT 1,
    created_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP       NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_categories_slug (slug),
    KEY idx_categories_active_sort (is_active, sort_order),
    KEY idx_categories_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    category_id         BIGINT UNSIGNED NOT NULL,
    slug                VARCHAR(160)    NOT NULL,
    pricing_mode        ENUM('unit','weight') NOT NULL DEFAULT 'unit',
    base_price_cents    BIGINT          NOT NULL DEFAULT 0, -- unit: base price; weight: price per kg
    is_preorder         TINYINT(1)      NOT NULL DEFAULT 0,
    preorder_lead_hours INT             NOT NULL DEFAULT 0,
    is_available        TINYINT(1)      NOT NULL DEFAULT 1,
    image_url           VARCHAR(500)    NULL,
    allergens_json      JSON            NULL, -- e.g. ["dairy","nuts","gluten"]
    sort_order          INT             NOT NULL DEFAULT 0,
    created_at          TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP       NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_products_slug (slug),
    KEY idx_products_category (category_id),
    KEY idx_products_available_sort (is_available, sort_order),
    KEY idx_products_deleted_at (deleted_at),
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Size variants for per-unit products (e.g. Kaak L 1.50 / S 1.00).
CREATE TABLE product_variants (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    product_id   BIGINT UNSIGNED NOT NULL,
    sku          VARCHAR(60)     NULL,
    label        VARCHAR(80)     NOT NULL, -- default-locale label ("Large"/"Small")
    price_cents  BIGINT          NOT NULL DEFAULT 0,
    is_default   TINYINT(1)      NOT NULL DEFAULT 0,
    sort_order   INT             NOT NULL DEFAULT 0,
    created_at   TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_variants_product (product_id),
    KEY idx_variants_deleted_at (deleted_at),
    CONSTRAINT fk_variants_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Option/add-on groups (e.g. Katayef filling: Kashta/Cheese/Walnuts).
CREATE TABLE modifier_groups (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    product_id  BIGINT UNSIGNED NOT NULL,
    label       VARCHAR(120)    NOT NULL,
    min_select  INT             NOT NULL DEFAULT 0,
    max_select  INT             NOT NULL DEFAULT 1,
    is_required TINYINT(1)      NOT NULL DEFAULT 0,
    sort_order  INT             NOT NULL DEFAULT 0,
    created_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_modgroups_product (product_id),
    KEY idx_modgroups_deleted_at (deleted_at),
    CONSTRAINT fk_modgroups_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Individual options within a group (e.g. Extra Cheese +2.00, Replace Akkawi +1.00).
CREATE TABLE modifiers (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    group_id          BIGINT UNSIGNED NOT NULL,
    label             VARCHAR(120)    NOT NULL,
    price_delta_cents BIGINT          NOT NULL DEFAULT 0,
    is_default        TINYINT(1)      NOT NULL DEFAULT 0,
    is_available      TINYINT(1)      NOT NULL DEFAULT 1,
    sort_order        INT             NOT NULL DEFAULT 0,
    created_at        TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at        TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_modifiers_group (group_id),
    KEY idx_modifiers_deleted_at (deleted_at),
    CONSTRAINT fk_modifiers_group FOREIGN KEY (group_id) REFERENCES modifier_groups (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Generic translation store (plan §4/§7). Holds per-locale field values for any
-- entity (category/product names & descriptions, etc.).
CREATE TABLE translations (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    entity_type VARCHAR(40)     NOT NULL, -- 'category' | 'product' | ...
    entity_id   BIGINT UNSIGNED NOT NULL,
    locale      VARCHAR(10)     NOT NULL, -- 'en' | 'ar' | 'fr'
    field       VARCHAR(40)     NOT NULL, -- 'name' | 'description'
    value       TEXT            NOT NULL,
    created_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_translation (entity_type, entity_id, locale, field),
    KEY idx_translation_lookup (entity_type, entity_id, locale)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
