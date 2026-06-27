-- Delivery zones (plan §5). A zone is either global (store-wide) or scoped to a
-- single product (a tighter radius for fragile/heavy items). Shape is radius
-- (center + radius_km, matched via Haversine) or polygon (GeoJSON ring, matched
-- via point-in-polygon).
CREATE TABLE delivery_zones (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name            VARCHAR(120)    NOT NULL,
    scope           ENUM('global','product') NOT NULL DEFAULT 'global',
    product_id      BIGINT UNSIGNED NULL,
    shape           ENUM('radius','polygon') NOT NULL DEFAULT 'radius',
    center_lat      DECIMAL(10,7)   NULL,
    center_lng      DECIMAL(10,7)   NULL,
    radius_km       DECIMAL(8,3)    NULL,
    polygon_geojson JSON            NULL,   -- [[lng,lat],...] ring
    fee_cents       BIGINT          NOT NULL DEFAULT 0,
    min_order_cents BIGINT          NOT NULL DEFAULT 0,
    is_active       TINYINT(1)      NOT NULL DEFAULT 1,
    sort_order      INT             NOT NULL DEFAULT 0,
    created_at      TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_zones_scope (scope, is_active),
    KEY idx_zones_product (product_id),
    KEY idx_zones_deleted_at (deleted_at),
    CONSTRAINT fk_zones_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
