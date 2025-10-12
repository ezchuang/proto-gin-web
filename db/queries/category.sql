-- name: CreateCategory :one
INSERT INTO category (name, slug) VALUES ($1, $2)
RETURNING id, name, slug;

-- name: GetCategoryBySlug :one
SELECT id, name, slug FROM category WHERE slug = $1;

-- name: ListCategories :many
SELECT id, name, slug FROM category ORDER BY name ASC LIMIT $1 OFFSET $2;

-- name: DeleteCategoryBySlug :exec
DELETE FROM category WHERE slug = $1;

