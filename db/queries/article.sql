-- name: CreateArticle :one
INSERT INTO article (title, body, author_id)
VALUES ($1, $2, $3)
RETURNING id, title, body, published_at, author_id, created_at;

-- name: GetArticle :one
SELECT id, title, body, published_at, author_id, created_at
FROM article
WHERE id = $1;

-- name: ListArticles :many
SELECT id, title, body, published_at, author_id, created_at
FROM article
ORDER BY id DESC
LIMIT $1 OFFSET $2;

