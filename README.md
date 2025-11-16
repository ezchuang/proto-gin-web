# Proto Gin Web (Blog Starter Template)

Go + Gin blog starter prototypes. It showcases clean layering (Clean-ish), sqlc + pgx typed queries, SSR + SEO hooks, modular usecases and routing, connection pooling, structured logging, and readiness checks.

## Core Features
- Go 1.24 (toolchain go1.25) with Gin router and middleware (Request-ID, structured logging, cache-control)
- PostgreSQL 16 (Docker), pgx v5 + sqlc typed query package, Flyway migrations
- SSR pages: home, post list, post detail with template helper `timefmt`
- SEO endpoints: `robots.txt`, `sitemap.xml`, `rss.xml` plus extendable SEO stubs
- Health probes: `/livez` for liveness, `/readyz` to verify DB connectivity
- Admin sample: Redis-backed admin sessions + remember-me, login/logout, post CRUD, category/tag management and relations

## Project Layout
```plaintext
proto─gin─web/
├─ cmd/
|   └─ api/main.go
├─ db/
|   ├─ migrations/
|   └─ queries/
├─ internal/
|   ├─ admin/
|   |   ├─ auth/{domain,adapters/http}
|   |   ├─ content/{app,adapters/http}
|   |   └─ ui/{app,adapters/http,adapters/view}
|   ├─ blog/
|   |   ├─ post/{domain,adapters/api,adapters/public,adapters/view}
|   |   └─ taxonomy/domain
|   ├─ infrastructure/
|   |   ├─ pg/
|   |   └─ platform/
|   ├─ platform/
|   |   ├─ http/{middleware,templates,view}
|   |   └─ seo/
|   └─ interfaces/
|       └─ auth/            (legacy cookie middleware)
├─ web/static/
├─ Dockerfile
├─ docker─compose.yml
├─ Makefile
└─ sqlc.yaml
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
4. 產生/更新 Swagger 文件
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   swag init -g cmd/api/main.go -o docs
   ```
   開發環境可透過 `http://localhost:8080/swagger/index.html` 查看文件（production 預設不開放）。

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
- Redis sessions: `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`
- App: `APP_ENV` (development/production), `PORT` (default 8080)
- Cookies: `ADMIN_SESSION_COOKIE`, `ADMIN_REMEMBER_COOKIE`
- SEO: `BASE_URL`, `SITE_NAME`, `SITE_DESCRIPTION`
- Compose: `HOST_POSTGRES_PORT`, `HOST_APP_PORT`, `HOST_REDIS_PORT`

### Sample Accounts
Database migrations seed two demo admins and both Argon2 hashes resolve to the plain text password `password`.

| Email               | Role  | Password |
|---------------------|-------|----------|
| `admin@example.com` | admin | `password` |
| `demo@example.com`  | admin | `password` |

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
- `internal/admin/*` + `internal/blog/*` follow context modules (`domain`, `app`, `adapters`, `view`) to keep DDD/CA seams explicit; legacy SSR admin UI is now housed in `internal/admin/ui/adapters/http`.
- `internal/platform/http` owns the router, middleware, templates, and now enforces Redis-backed admin sessions plus remember-me recovery.
- `internal/infrastructure/{pg,platform}` provide persistence implementations and config/logging bootstrap shared across contexts.
- `db/queries` + `sqlc`: SQL -> typed accessors
- Observability: Request-ID middleware, structured slog logging, cache-control helper, readiness probe

## Roadmap Ideas
- Session hardening: device metadata, audit logs, admin session management UI
- Editorial workflow & versioning (post revision workflow)
- Full-text search (tsvector)
- Asset pipeline: fingerprint + minify static files, tighten CDN/Nginx caching
- Observability: metrics, tracing, error tracking
- Alternative persistence (e.g. GORM layer) for comparison
