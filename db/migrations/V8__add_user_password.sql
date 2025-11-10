-- Add password hash column to app_user and seed default admin argon2 hash

ALTER TABLE app_user
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

UPDATE app_user
SET password_hash = $$$argon2id$v=19$m=65536,t=2,p=1$v9QLeR8Xy34zeByEU/5wFw$T1Z2bsXDSvdsVwMpz1iC8Nf1F9uwM88a43Uzinn6d9c$$
WHERE email = 'admin@example.com';

ALTER TABLE app_user
    ALTER COLUMN password_hash SET NOT NULL;
