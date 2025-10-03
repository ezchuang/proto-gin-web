# Proto Gin Web（企業用部落格原型）

以 Go + Gin 打造的輕量「企業用部落格」原型，強調清晰分層（Clean-ish）、型別安全 SQL（sqlc + pgx）、SSR 與 SEO 基礎。適合學習、面試與快速起手。

## 技術棧
- Go 1.24（toolchain go1.25）、Gin（router/middleware/SSR templates）
- PostgreSQL 16（Docker）、pgx v5、sqlc
- Flyway（migrations）
- SEO：Robots.txt / Sitemap.xml / RSS.xml、Open Graph、JSON-LD（stub）
- slog（JSON log）、godotenv（本機載入 .env）
- Docker Compose（db/api/flyway）

## 目錄重點
proto-gin-web/
```plaintext
.
├── cmd/
│   └── api/
│       └── main.go
├── db/
│   ├── migrations/
│   │   ├── V1init.sql
│   │   ├── V2blog.sql
│   │   └── V3indexes.sql
│   └── queries/ # sqlc SQL 檔
│       ├── article.sql
│       ├── category.sql
│       ├── post_relation.sql
│       ├── post.sql
│       └── tag.sql
├── internal/
│   ├── auth/
│   │   └── session.go # admin cookie stub
│   ├── core/ # entities & usecases (thin)
│   │   ├── post.go
│   │   └── post_service.go
│   ├── http/ # router, handlers, templates
│   │   └── views/
│   │   │   ├── index.tmpl
│   │   │   ├── layout.tmpl
│   │   │   └── post.tmpl
│   │   ├── handlers.go
│   │   └── router.go
│   ├── platform/ # config, logger
│   │   ├── config.go
│   │   └── logger.go
│   ├── repo/
│   │   └── pg/
│   │       └── db.go # pgx pool
│   └── seo/ # seo utils (stubs)
│       └── meta.go
│       ├── rss.go
│       └── sitemap.go
├── web/
│   └── static/
│       └── app.css
├── sqlc.yaml
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
```

## 快速開始
1) 啟 DB 並套用遷移
```bash
make db-up
make migrate
```
2) 產生 sqlc 程式碼（擇一）
```bash
make sqlc          # 需本機安裝 sqlc
# 或
make sqlc-docker   # 容器產生
```
3) 下載依賴與啟動
```bash
make deps
make run           # http://localhost:8080
```

### 容器化
```bash
make up            # 啟動 API + DB（先 make migrate 一次）
make logs
make down
```

## Makefile 常用
- db：`db-up`、`db-psql`、`migrate`、`migrate-info`、`migrate-repair`
- 產碼：`sqlc`、`sqlc-docker`
- 應用：`deps`、`run`、`build`、`up`、`logs`、`down`

## 環境變數
- DB：`POSTGRES_USER`、`POSTGRES_PASSWORD`、`POSTGRES_DB`、`POSTGRES_HOST`、`POSTGRES_PORT`
- App：`APP_ENV`（development/production）、`PORT`（預設 8080）
- SEO：`BASE_URL`、`SITE_NAME`、`SITE_DESCRIPTION`
- Admin：`ADMIN_USER`、`ADMIN_PASS`
- Compose：`HOST_POSTGRES_PORT`、`HOST_APP_PORT`

## 功能總覽
- SSR：首頁 `/`、列表 `/posts?page=1&size=10&category=&tag=&sort=`、單篇 `/posts/:slug`（Markdown 安全渲染）
- SEO：`/robots.txt`、`/sitemap.xml`（DB 生成）、`/rss.xml`（DB 生成）
- API：
  - `GET /api/articles?limit=10&offset=0`
  - `GET /api/posts?category=...&tag=...&sort=created_at_desc|created_at_asc|published_at_desc|published_at_asc&limit=&offset=`
- Admin（最小可用）
  - 登入/登出：`POST /admin/login`、`POST /admin/logout`
  - 貼文：`POST /admin/posts`、`PUT /admin/posts/:slug`、`DELETE /admin/posts/:slug`
  - 類別：`POST /admin/categories`、`DELETE /admin/categories/:slug`
  - 標籤：`POST /admin/tags`、`DELETE /admin/tags/:slug`
  - 關聯：
    - `POST /admin/posts/:slug/categories/:cat`、`DELETE /admin/posts/:slug/categories/:cat`
    - `POST /admin/posts/:slug/tags/:tag`、`DELETE /admin/posts/:slug/tags/:tag`

## 開發說明（Clean-ish）
- `http`：HTTP concern（router/middleware/SSR），商業邏輯盡量往 `core` 下沉
- `repo/pg`：連線管理；SQL 由 `sqlc` 產出於 `internal/data`
- `platform`：設定與日誌
- SQL：修改 `db/queries/*.sql` 後執行 `make sqlc`
- 遷移：新增 `db/migrations/V*.sql` 後 `make migrate`

## Roadmap 建議
- 後台頁面與完整 session（JWT 或 server-side session）
- 內容審核/版本（`post_revision`）、草稿工作流
- 列表過濾/排序更完整、全文搜尋（tsvector）
- 靜態資源優化與快取策略（ETag/Last-Modified）
- 可觀測性（metrics/tracing/error tracking）
- GORM 可替換層示範