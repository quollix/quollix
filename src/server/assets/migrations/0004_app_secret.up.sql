CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS app_secret TEXT;

UPDATE apps
SET app_secret = encode(gen_random_bytes(32), 'hex')
WHERE app_secret IS NULL
   OR app_secret = '';

ALTER TABLE apps
    ALTER COLUMN app_secret SET NOT NULL;

ALTER TABLE apps
    ADD CONSTRAINT apps_app_secret_not_empty CHECK (app_secret <> '');
