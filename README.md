# Proto Gin Web

Prototype blog/admin backend built with Go + Gin. The goal is to provide a pragmatic template that shows how to wire clean-ish layering, sqlc/pgx repositories, Redis-backed admin sessions (with remember-me), modular routing, Swagger, and Docker-first workflows. Use it as a reference or seed for small/medium services.

---

## Highlights

- **Go 1.24 / toolchain go1.25** with Gin router, middleware (Request-ID, slog logger, recovery, IP rate limiter) and HTML templates.
- **PostgreSQL 16 + sqlc + pgx v5** for typed queries, repository interfaces, and Flyway migrations.
- **Redis session store + remember-me tokens** guarded by Argon2id passwords, cookie hardening, and session/device revocation helpers.
- **Clean architecture contexts** (`admin`, `blog`, `platform`, `infrastructure`) keeping domain/app/adapters boundaries explicit.
- **Public + admin surfaces**: SEO routes (`robots.txt`, `sitemap.xml`, `rss.xml`), `/livez` + `/readyz`, `/api/posts`, admin JSON APIs, and legacy admin UI.
- **Dev experience**: Make targets, Docker Compose services (Postgres, Redis, API), swagger generation, and structured logging to stdout/file.

---

## Architecture Overview

| Context | Responsibilities | Key Paths |
|---------|------------------|-----------|
| `internal/admin` | Auth (login/register/profile), content CRUD (posts/categories/tags), legacy admin UI | `internal/admin/auth`, `internal/admin/content`, `internal/admin/ui` |
| `internal/application` | Use cases/interfaces for admin, post, taxonomy domains | `internal/application/{admin,post,taxonomy}` |
| `internal/blog` | Public pages + API + SEO + taxonomy models | `internal/blog/post`, `internal/blog/taxonomy` |
| `internal/infrastructure` | pgx repositories, Redis session store, platform config/logger/feed helpers | `internal/infrastructure/{pg,redis,platform,feed}` |
| `internal/platform/http` | Router, middleware, template loader, responder, SEO helpers | `internal/platform/http` |

### Directory Layout (excerpt)

```
proto-gin-web/
├─ cmd/api/main.go
├─ db/{migrations,queries}
├─ internal/
│  ├─ admin/{auth,content,ui}
│  ├─ application/{admin,post,taxonomy}
│  ├─ blog/post/{domain,adapters}
│  ├─ infrastructure/{pg,redis,platform}
│  └─ platform/http/{middleware,templates,view,seo}
├─ docs/          # swag output
├─ web/static/    # css + demo assets + uploads
├─ Dockerfile
├─ docker-compose.yml
├─ Makefile
└─ sqlc.yaml
```

---

## Features

### Public
- `GET /` landing page, `GET /posts` (pagination/filter/sort), `GET /posts/:slug`.
- SEO: `GET /robots.txt`, `GET /sitemap.xml`, `GET /rss.xml`.
- Health probes: `GET /livez`, `GET /readyz`.
- JSON API: `GET /api/posts?limit=&offset=&category=&tag=&sort=`, `GET /api/posts/:slug`.

### Admin API
- Auth: `POST /admin/login`, `POST /admin/logout`, `POST /admin/register`, `GET/POST /admin/profile`.
- Content: `POST /admin/posts`, `PUT /admin/posts/:slug`, `DELETE /admin/posts/:slug`.
- Taxonomy: `POST /admin/categories`, `DELETE /admin/categories/:slug`, `POST /admin/tags`, `DELETE /admin/tags/:slug`.
- Relations: `POST/DELETE /admin/posts/:slug/categories/:cat`, `POST/DELETE /admin/posts/:slug/tags/:tag`.

### Security & Observability
- Argon2id hashing, input normalization, remember-me split tokens stored in Postgres.
- Redis session store with per-user session sets, TTL refresh (`Touch`), and device revocation.
- Middleware stack: security headers, Request-ID, slog-based logger, panic recovery, rate limiter.
- Swagger docs available under `/swagger/index.html` in non-production envs.

---

## Getting Started

### Prerequisites
- Go 1.24+ (toolchain go1.25 recommended)
- Docker + Docker Compose

### Local Run
```bash
make db-up          # start postgres
make migrate        # flyway migrations
make sqlc-docker    # or make sqlc
make deps
make run            # http://localhost:8080
```

### Docker Workflow
```bash
make up             # build+start api + db
make logs           # tail api/db logs
make down           # stop & remove volumes
```

### Swagger Docs
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/api/main.go -o docs
# visit http://localhost:8080/swagger/index.html (non-production)
```

---

## Environment Variables

| Category | Variables |
|----------|-----------|
| Database | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_HOST`, `POSTGRES_PORT` |
| Redis    | `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` |
| App      | `APP_ENV` (`development`/`production`), `PORT` (default `8080`), `BASE_URL`, `SITE_NAME`, `SITE_DESCRIPTION` |
| Cookies  | `ADMIN_SESSION_COOKIE`, `ADMIN_REMEMBER_COOKIE` |
| Compose  | `HOST_POSTGRES_PORT`, `HOST_APP_PORT`, `HOST_REDIS_PORT` |

Configure via `.env` (copy `.env.example`) or environment overrides.

---

## Sample Admin Accounts

Database seeds create two demo admins (password = `password`):

| Email               | Role  |
|---------------------|-------|
| `admin@example.com` | admin |
| `demo@example.com`  | admin |

Use them for UI/API testing.

---

## Make Targets Cheat Sheet

| Category | Targets |
|----------|---------|
| Database | `db-up`, `db-psql`, `migrate`, `migrate-info`, `migrate-repair`, `db-down` |
| Codegen  | `sqlc`, `sqlc-docker` |
| App      | `deps`, `run`, `build`, `up`, `logs`, `down`, `api-build`, `clean` |

---

## Roadmap Ideas

- Admin session hardening: device metadata, audit logs, self-service session management.
- Editorial workflow: drafts, revisions, publishing queues.
- Search: PostgreSQL full-text or external search service.
- Asset pipeline: hashing/minify static files, CDN cache hints.
- Enhanced observability: metrics, tracing, error tracking integrations.
- Alternate persistence layer (e.g., GORM) for comparison/testing.

---

MIT Licensed. Use, fork, and adapt as needed.
