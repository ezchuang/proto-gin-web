-- Persist remember-me split tokens for admin accounts

CREATE TABLE IF NOT EXISTS admin_remember_tokens (
    selector TEXT PRIMARY KEY,
    validator_hash TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
    device_info TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_admin_remember_tokens_user_id
    ON admin_remember_tokens (user_id);

CREATE INDEX IF NOT EXISTS idx_admin_remember_tokens_expires_at
    ON admin_remember_tokens (expires_at);
