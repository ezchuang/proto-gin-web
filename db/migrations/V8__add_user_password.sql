-- Add password hash column to app_user and seed default admin hash

ALTER TABLE app_user
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

UPDATE app_user
SET password_hash = $$2a$10$7EqJtq98hPqEX7fNZaFWoOhi/PrjwIEmS8Zk8KfDeFZZzI6Fz3p4a$$
WHERE password_hash IS NULL;

ALTER TABLE app_user
    ALTER COLUMN password_hash SET NOT NULL;
