-- Helpful indexes for blog queries and relations

CREATE INDEX IF NOT EXISTS idx_post_published_at ON post (published_at);
CREATE INDEX IF NOT EXISTS idx_post_created_at ON post (created_at);
CREATE INDEX IF NOT EXISTS idx_post_status ON post (status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_post_slug ON post (slug);

CREATE UNIQUE INDEX IF NOT EXISTS idx_category_slug ON category (slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_slug ON tag (slug);

CREATE INDEX IF NOT EXISTS idx_post_category_post ON post_category (post_id);
CREATE INDEX IF NOT EXISTS idx_post_category_category ON post_category (category_id);
CREATE INDEX IF NOT EXISTS idx_post_tag_post ON post_tag (post_id);
CREATE INDEX IF NOT EXISTS idx_post_tag_tag ON post_tag (tag_id);

