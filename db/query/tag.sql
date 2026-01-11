-- name: CreateTag :one
INSERT INTO tags (name, slug, color, category_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListTags :many
SELECT *
FROM tags
ORDER BY name;

