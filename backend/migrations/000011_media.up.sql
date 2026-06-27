-- Media: uploaded images with optimized + thumbnail variants.
-- image_url / thumb_url store the public-facing absolute URLs so catalog
-- entities can simply reference them without re-resolving storage keys.

CREATE TABLE media (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    filename      VARCHAR(200)    NOT NULL,            -- stored name (uuid.jpg)
    original_name VARCHAR(255)    NOT NULL,            -- user's original filename
    mime          VARCHAR(80)     NOT NULL,            -- verified MIME type
    ext           VARCHAR(10)     NOT NULL,            -- file extension
    size_bytes    BIGINT          NOT NULL DEFAULT 0,  -- uploaded file size in bytes
    width         INT             NOT NULL DEFAULT 0,  -- optimized image width px
    height        INT             NOT NULL DEFAULT 0,  -- optimized image height px
    url           VARCHAR(500)    NOT NULL,            -- public URL of optimized original
    thumb_url     VARCHAR(500)    NOT NULL,            -- public URL of thumbnail
    alt           VARCHAR(255)    NULL,                -- optional alt text / caption
    created_by    BIGINT UNSIGNED NULL,                -- admin who uploaded (FK users)
    created_at    TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP       NULL,
    PRIMARY KEY (id),
    KEY idx_media_created_at (created_at),
    KEY idx_media_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
