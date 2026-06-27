DELETE FROM settings WHERE `key` IN (
    'square_environment','square_application_id','square_location_id',
    'square_access_token','square_webhook_signature_key'
);
DROP TABLE IF EXISTS payment_transactions;
