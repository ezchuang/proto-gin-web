-- name: AddCategoryToPost :exec
INSERT INTO post_category (post_id, category_id)
SELECT p.id, c.id FROM post p, category c WHERE p.slug = $1 AND c.slug = $2
ON CONFLICT DO NOTHING;

-- name: RemoveCategoryFromPost :exec
DELETE FROM post_category USING post p, category c
WHERE post_category.post_id = p.id AND post_category.category_id = c.id
  AND p.slug = $1 AND c.slug = $2;

-- name: AddTagToPost :exec
INSERT INTO post_tag (post_id, tag_id)
SELECT p.id, t.id FROM post p, tag t WHERE p.slug = $1 AND t.slug = $2
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromPost :exec
DELETE FROM post_tag USING post p, tag t
WHERE post_tag.post_id = p.id AND post_tag.tag_id = t.id
  AND p.slug = $1 AND t.slug = $2;

-- name: ListCategoriesByPostSlug :many
SELECT c.id, c.name, c.slug
FROM category c
JOIN post_category pc ON pc.category_id = c.id
JOIN post p ON p.id = pc.post_id
WHERE p.slug = $1
ORDER BY c.name ASC;

-- name: ListTagsByPostSlug :many
SELECT t.id, t.name, t.slug
FROM tag t
JOIN post_tag pt ON pt.tag_id = t.id
JOIN post p ON p.id = pt.post_id
WHERE p.slug = $1
ORDER BY t.name ASC;

