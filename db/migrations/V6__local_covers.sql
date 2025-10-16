-- Switch covers to local /static assets for offline demos
UPDATE post SET cover_url = '/static/cover-gin.svg' WHERE slug = 'getting-started-gin';
UPDATE post SET cover_url = '/static/cover-ssr.svg' WHERE slug = 'ssr-with-gin';
UPDATE post SET cover_url = '/static/cover-sqlc.svg' WHERE slug = 'pgx-sqlc-typed-queries';
UPDATE post SET cover_url = '/static/cover-docker.svg' WHERE slug = 'dockerizing-gin-app';
UPDATE post SET cover_url = '/static/cover-paging.svg' WHERE slug = 'pagination-sorting-filtering';
UPDATE post SET cover_url = '/static/cover-seo.svg' WHERE slug = 'seo-sitemap-rss';
UPDATE post SET cover_url = '/static/cover-testing.svg' WHERE slug = 'testing-handlers-services';
UPDATE post SET cover_url = '/static/cover-release.svg' WHERE slug = 'release-notes-v0-1';

