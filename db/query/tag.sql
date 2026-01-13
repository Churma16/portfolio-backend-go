-- name: CreateTag :one
INSERT INTO tags (name, slug, color, category_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetTags :many
SELECT *
FROM tags
ORDER BY name;

-- name: GetTag :one
SELECT *
FROM tags
WHERE id = $1 LIMIT 1;

-- name: UpdateTag :one
UPDATE tags
SET name        = $2,
    slug        = $3,
    color       = $4,
    category_id = $5,
    updated_at  = now()
WHERE id = $1 RETURNING *;

-- name: DeleteTag :one
DELETE
FROM tags
WHERE id = $1 RETURNING *;

