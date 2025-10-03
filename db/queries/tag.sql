-- name: CreateTag :one
INSERT INTO tag (name, slug) VALUES ($1, $2)
RETURNING id, name, slug;

-- name: GetTagBySlug :one
SELECT id, name, slug FROM tag WHERE slug = $1;

-- name: ListTags :many
SELECT id, name, slug FROM tag ORDER BY name ASC LIMIT $1 OFFSET $2;

-- name: DeleteTagBySlug :exec
DELETE FROM tag WHERE slug = $1;

