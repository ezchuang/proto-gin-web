-- Init

CREATE TABLE IF NOT EXISTS app_user (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT NOT NULL,
    body          TEXT NOT NULL,
    published_at  TIMESTAMPTZ,
    author_id     BIGINT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed a demo user and article for quick checks
INSERT INTO app_user (email, display_name)
VALUES ('demo@example.com', 'Demo User')
ON CONFLICT (email) DO NOTHING;

INSERT INTO article (title, body, author_id)
SELECT 'Hello Proto', 'This is your first article.', id FROM app_user WHERE email = 'demo@example.com'
ON CONFLICT DO NOTHING;

