-- Seed richer example content: categories, tags, posts, relations

-- Categories
INSERT INTO category (name, slug) VALUES
  ('Guides', 'guides'),
  ('Go', 'go'),
  ('DevOps', 'devops'),
  ('Architecture', 'architecture'),
  ('News', 'news'),
  ('Tutorials', 'tutorials')
ON CONFLICT (slug) DO NOTHING;

-- Tags
INSERT INTO tag (name, slug) VALUES
  ('gin', 'gin'),
  ('sqlc', 'sqlc'),
  ('pgx', 'pgx'),
  ('postgres', 'postgres'),
  ('docker', 'docker'),
  ('testing', 'testing'),
  ('performance', 'performance'),
  ('security', 'security'),
  ('routing', 'routing'),
  ('templates', 'templates'),
  ('release', 'release')
ON CONFLICT (slug) DO NOTHING;

-- Helper: author id for admin
-- (Flyway runs this whole file in a single transaction by default)
WITH author AS (
  SELECT id FROM app_user WHERE email = 'admin@example.com'
)
-- Posts (use different publish times)
INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Getting Started with Gin', 'getting-started-gin',
       'Quickstart guide to building APIs with Gin.',
       '# Getting Started\n\nThis post walks through setting up a minimal Gin app, routing, and middleware.\n\n## Steps\n- Install Go and Gin\n- Create a new module\n- Add a simple route\n\n```go\nr := gin.Default()\nr.GET("/ping", func(c *gin.Context){ c.JSON(200, gin.H{""message"": ""pong""}) })\n```',
       'https://picsum.photos/seed/gin/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '10 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Server-Side Rendering with Gin', 'ssr-with-gin',
       'Render server-side HTML templates safely with Gin.',
       '# SSR with Gin\n\nUse Go templates to render pages. Sanitize any user content before rendering.\n\n- Layouts and partials\n- Template functions\n- Safe HTML',
       'https://picsum.photos/seed/ssr/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '9 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Typed Postgres with pgx + sqlc', 'pgx-sqlc-typed-queries',
       'Compile-time safe queries using pgx and sqlc.',
       '# pgx + sqlc\n\nLeverage `sqlc` to generate type-safe accessors over `pgx`.\n\n- Define queries in `db/queries`\n- Generate code\n- Use in repositories',
       'https://picsum.photos/seed/sqlc/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '8 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Dockerizing Your Gin App', 'dockerizing-gin-app',
       'Containerize your app and wire migrations.',
       '# Docker + Compose\n\nUse Docker Compose for Postgres and your API. Run Flyway for migrations.',
       'https://picsum.photos/seed/docker/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '7 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Pagination, Sorting, and Filtering', 'pagination-sorting-filtering',
       'Build a clean list view with paging, sort, and filters.',
       '# Lists that scale\n\nSupport `page`, `size`, and taxonomy filters with proper indexes.',
       'https://picsum.photos/seed/paging/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '6 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'SEO Basics: Sitemap and RSS', 'seo-sitemap-rss',
       'Expose sitemap.xml and rss.xml for better discovery.',
       '# SEO hooks\n\nAdd `/sitemap.xml` and `/rss.xml` endpoints sourcing from published posts.',
       'https://picsum.photos/seed/seo/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '5 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Testing Handlers and Services', 'testing-handlers-services',
       'Unit and integration tests for confidence.',
       '# Testing strategy\n\nUse table-driven tests and lightweight integration tests against Postgres.',
       NULL,
       'published', (SELECT id FROM author), NOW() - INTERVAL '4 days'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
SELECT 'Release Notes: v0.1 Demo', 'release-notes-v0-1',
       'Highlights of the initial demo release.',
       '# v0.1\n\nInitial public demo with SSR, sqlc, and basic SEO.',
       'https://picsum.photos/seed/release/1200/600',
       'published', (SELECT id FROM author), NOW() - INTERVAL '3 days'
ON CONFLICT (slug) DO NOTHING;

-- A draft that should not appear on /posts
INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id)
SELECT 'Work In Progress', 'work-in-progress',
       'Draft content not yet published.',
       '# WIP\n\nThis is a draft and should not show up in listings.',
       NULL,
       'draft', (SELECT id FROM app_user WHERE email = 'admin@example.com')
ON CONFLICT (slug) DO NOTHING;

-- Relations: categories and tags per post
-- Getting Started with Gin
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'getting-started-gin' AND c.slug IN ('guides', 'tutorials')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'getting-started-gin' AND t.slug IN ('gin', 'routing')
ON CONFLICT DO NOTHING;

-- SSR with Gin
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'ssr-with-gin' AND c.slug IN ('guides', 'architecture')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'ssr-with-gin' AND t.slug IN ('gin', 'templates')
ON CONFLICT DO NOTHING;

-- pgx + sqlc
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'pgx-sqlc-typed-queries' AND c.slug IN ('go')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'pgx-sqlc-typed-queries' AND t.slug IN ('pgx', 'sqlc', 'postgres')
ON CONFLICT DO NOTHING;

-- Dockerizing Your Gin App
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'dockerizing-gin-app' AND c.slug IN ('devops')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'dockerizing-gin-app' AND t.slug IN ('docker')
ON CONFLICT DO NOTHING;

-- Pagination, Sorting, and Filtering
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'pagination-sorting-filtering' AND c.slug IN ('guides')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'pagination-sorting-filtering' AND t.slug IN ('routing', 'performance')
ON CONFLICT DO NOTHING;

-- SEO Basics: Sitemap and RSS
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'seo-sitemap-rss' AND c.slug IN ('architecture')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'seo-sitemap-rss' AND t.slug IN ('templates', 'routing')
ON CONFLICT DO NOTHING;

-- Testing Handlers and Services
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'testing-handlers-services' AND c.slug IN ('go')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'testing-handlers-services' AND t.slug IN ('testing')
ON CONFLICT DO NOTHING;

-- Release Notes
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = 'release-notes-v0-1' AND c.slug IN ('news')
ON CONFLICT DO NOTHING;
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = 'release-notes-v0-1' AND t.slug IN ('release')
ON CONFLICT DO NOTHING;

