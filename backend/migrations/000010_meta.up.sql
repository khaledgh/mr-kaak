-- Meta (Facebook) Pixel + Conversions API settings (plan §14). Drop in keys to
-- activate; no code change needed.
INSERT INTO settings (`key`, value_json, description) VALUES
    ('meta_pixel_id',        '""', 'Meta Pixel ID (frontend)'),
    ('meta_capi_token',      '""', 'Meta Conversions API access token (secret)'),
    ('meta_test_event_code', '""', 'Meta test event code (optional)')
ON DUPLICATE KEY UPDATE description = VALUES(description);
