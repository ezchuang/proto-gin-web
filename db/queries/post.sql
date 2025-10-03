-- name: CreatePost :one
INSERT INTO post (title, slug, summary, content_md, cover_url, status, author_id, published_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at;

-- name: GetPostBySlug :one
SELECT id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at
FROM post
WHERE slug = $1;

-- name: ListPublishedPosts :many
SELECT id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at
FROM post
WHERE status = 'published'
ORDER BY COALESCE(published_at, created_at) DESC
LIMIT $1 OFFSET $2;

-- name: ListPublishedPostsByCategory :many
SELECT p.id, p.title, p.slug, p.summary, p.content_md, p.cover_url, p.status, p.author_id, p.published_at, p.created_at, p.updated_at
FROM post p
JOIN post_category pc ON pc.post_id = p.id
JOIN category c ON c.id = pc.category_id
WHERE p.status = 'published' AND c.slug = $1
ORDER BY COALESCE(p.published_at, p.created_at) DESC
LIMIT $2 OFFSET $3;

-- name: ListPublishedPostsByTag :many
SELECT p.id, p.title, p.slug, p.summary, p.content_md, p.cover_url, p.status, p.author_id, p.published_at, p.created_at, p.updated_at
FROM post p
JOIN post_tag pt ON pt.post_id = p.id
JOIN tag t ON t.id = pt.tag_id
WHERE p.status = 'published' AND t.slug = $1
ORDER BY COALESCE(p.published_at, p.created_at) DESC
LIMIT $2 OFFSET $3;

-- name: ListPublishedPostsSorted :many
SELECT id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at
FROM post
WHERE status = 'published'
ORDER BY
  CASE WHEN $1 = 'published_at_asc' THEN published_at END ASC,
  CASE WHEN $1 = 'published_at_desc' THEN published_at END DESC,
  CASE WHEN $1 = 'created_at_asc' THEN created_at END ASC,
  CASE WHEN $1 = 'created_at_desc' OR $1 = '' THEN created_at END DESC
NULLS LAST
LIMIT $2 OFFSET $3;

-- name: ListPublishedPostsByCategorySorted :many
SELECT p.id, p.title, p.slug, p.summary, p.content_md, p.cover_url, p.status, p.author_id, p.published_at, p.created_at, p.updated_at
FROM post p
JOIN post_category pc ON pc.post_id = p.id
JOIN category c ON c.id = pc.category_id
WHERE p.status = 'published' AND c.slug = $1
ORDER BY
  CASE WHEN $2 = 'published_at_asc' THEN p.published_at END ASC,
  CASE WHEN $2 = 'published_at_desc' THEN p.published_at END DESC,
  CASE WHEN $2 = 'created_at_asc' THEN p.created_at END ASC,
  CASE WHEN $2 = 'created_at_desc' OR $2 = '' THEN p.created_at END DESC
NULLS LAST
LIMIT $3 OFFSET $4;

-- name: ListPublishedPostsByTagSorted :many
SELECT p.id, p.title, p.slug, p.summary, p.content_md, p.cover_url, p.status, p.author_id, p.published_at, p.created_at, p.updated_at
FROM post p
JOIN post_tag pt ON pt.post_id = p.id
JOIN tag t ON t.id = pt.tag_id
WHERE p.status = 'published' AND t.slug = $1
ORDER BY
  CASE WHEN $2 = 'published_at_asc' THEN p.published_at END ASC,
  CASE WHEN $2 = 'published_at_desc' THEN p.published_at END DESC,
  CASE WHEN $2 = 'created_at_asc' THEN p.created_at END ASC,
  CASE WHEN $2 = 'created_at_desc' OR $2 = '' THEN p.created_at END DESC
NULLS LAST
LIMIT $3 OFFSET $4;

-- name: UpdatePostBySlug :one
UPDATE post
SET title = $2,
    summary = $3,
    content_md = $4,
    cover_url = $5,
    status = $6,
    updated_at = NOW()
WHERE slug = $1
RETURNING id, title, slug, summary, content_md, cover_url, status, author_id, published_at, created_at, updated_at;

-- name: DeletePostBySlug :exec
DELETE FROM post WHERE slug = $1;

