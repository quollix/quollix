-- Synthetic example.invalid emails are reserved per username, so old mismatches
-- are normalized before restoring the unique email constraint.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

UPDATE users
SET email = username || '@example.invalid'
WHERE email LIKE '%@example.invalid'
  AND email <> username || '@example.invalid';

-- Admins can disable user accounts.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_enabled BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
