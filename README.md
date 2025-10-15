# Proto Gin Web (Blog Starter Template)

Go + Gin blog starter prototypes. It showcases clean layering (Clean-ish), sqlc + pgx typed queries, SSR + SEO hooks, modular usecases and routing, connection pooling, structured logging, and readiness checks.

## Core Features
- Go 1.24 (toolchain go1.25) with Gin router and middleware (Request-ID, structured logging, cache-control)
- PostgreSQL 16 (Docker), pgx v5 + sqlc typed query package, Flyway migrations
- SSR pages: home, post list, post detail with template helper `timefmt`
- SEO endpoints: `robots.txt`, `sitemap.xml`, `rss.xml` plus extendable SEO stubs
- Health probes: `/livez` for liveness, `/readyz` to verify DB connectivity
- Admin sample: login/logout, post CRUD, category/tag management and relations

## Project Layout
```plaintext
proto-gin-web/
├── cmd/
│   └── api/main.go
├── db/
│   ├── migrations/
│   │   ├── V1__init.sql
│   │   ├── V2__blog.sql
│   │   ├── V3__indexes.sql
│   │   └── V4__drop_article.sql
│   └── queries/
│       ├── category.sql
│       ├── post.sql
│       ├── post_relation.sql
│       └── tag.sql
├── internal/
│   ├── application/post/
│   ├── domain/
│   ├── infrastructure/
│   │   ├── pg/
│   │   ├── platform/
│   │   └── seo/
│   └── interfaces/http/
├── web/static/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── sqlc.yaml
```

## Getting Started
1. Start database & run migrations
   ```bash
   make db-up
   make migrate
   ```
2. Generate sqlc code
   ```bash
   make sqlc-docker    # or `make sqlc`
   ```
3. Install dependencies & run the app
   ```bash
   make deps
   make run            # http://localhost:8080
   ```

### Docker Workflow
```bash
make up
make logs
make down
```

## Useful Make Targets
- db: `db-up`, `db-psql`, `migrate`, `migrate-info`, `migrate-repair`
- codegen: `sqlc`, `sqlc-docker`
- app lifecycle: `deps`, `run`, `build`, `up`, `logs`, `down`

## Environment Variables
- DB: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_HOST`, `POSTGRES_PORT`
- App: `APP_ENV` (development/production), `PORT` (default 8080)
- SEO: `BASE_URL`, `SITE_NAME`, `SITE_DESCRIPTION`
- Admin: `ADMIN_USER`, `ADMIN_PASS`
- Compose: `HOST_POSTGRES_PORT`, `HOST_APP_PORT`

## Feature Summary
- SSR templates: `/`, `/posts` (pagination/filter/sort), `/posts/:slug`
- SEO: `/robots.txt`, `/sitemap.xml`, `/rss.xml`
- API: `GET /api/posts?category=&tag=&sort=&limit=&offset=`
- Admin sample:
  - Auth: POST `/admin/login`, POST `/admin/logout`
  - Posts: POST `/admin/posts`, PUT `/admin/posts/:slug`, DELETE `/admin/posts/:slug`
  - Categories: POST `/admin/categories`, DELETE `/admin/categories/:slug`
  - Tags: POST `/admin/tags`, DELETE `/admin/tags/:slug`
  - Relations: POST/DELETE `/admin/posts/:slug/categories/:cat`, POST/DELETE `/admin/posts/:slug/tags/:tag`

## Architecture Notes
- `internal/interfaces/http`: separates public/API/admin routes and middleware
- `internal/domain`: core entities and aggregate contracts
- `internal/application/post`: validation, normalization, taxonomy orchestration
- `internal/infrastructure/pg`: `pgx/sqlc` persistence implementation
- `internal/infrastructure/platform`: configuration & logging bootstrap
- `db/queries` + `sqlc`: SQL -> typed accessors
- Observability: Request-ID middleware, structured slog logging, cache-control helper, readiness probe

## Roadmap Ideas
- Production-ready auth/session (JWT or server-side sessions)
- Editorial workflow & versioning (post revision workflow)
- Full-text search (tsvector)
- Asset pipeline: fingerprint + minify static files, tighten CDN/Nginx caching
- Observability: metrics, tracing, error tracking
- Alternative persistence (e.g. GORM layer) for comparison
