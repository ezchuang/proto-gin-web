-- Enterprise blog schema (posts, categories, tags, users, roles)

CREATE TABLE IF NOT EXISTS role (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS app_user (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    role_id       BIGINT REFERENCES role(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS post (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT NOT NULL,
    slug          TEXT NOT NULL UNIQUE,
    summary       TEXT NOT NULL DEFAULT '',
    content_md    TEXT NOT NULL,
    cover_url     TEXT,
    status        TEXT NOT NULL DEFAULT 'draft', -- draft|published|archived
    author_id     BIGINT NOT NULL REFERENCES app_user(id) ON DELETE RESTRICT,
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS category (
    id    BIGSERIAL PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE,
    slug  TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tag (
    id    BIGSERIAL PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE,
    slug  TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS post_category (
    post_id     BIGINT NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES category(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);

CREATE TABLE IF NOT EXISTS post_tag (
    post_id BIGINT NOT NULL REFERENCES post(id) ON DELETE CASCADE,
    tag_id  BIGINT NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

-- Seed minimal roles and author
INSERT INTO role (name) VALUES ('admin') ON CONFLICT (name) DO NOTHING;
INSERT INTO app_user (email, display_name, role_id)
SELECT 'admin@example.com', 'Admin', r.id FROM role r WHERE r.name = 'admin'
ON CONFLICT (email) DO NOTHING;

-- Seed a sample published post
INSERT INTO post (title, slug, summary, content_md, status, author_id, published_at)
SELECT 'Welcome to Proto Blog', 'welcome', 'First enterprise blog post', '# Hello World', 'published', u.id, NOW()
FROM app_user u WHERE u.email = 'admin@example.com'
ON CONFLICT (slug) DO NOTHING;

