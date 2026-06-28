-- name: CreateTechStackCategory :one
INSERT INTO tech_stack_categories (name, slug, color, created_at, updated_at)
VALUES ($1, $2, $3, now(), now()) RETURNING *;

-- name: GetTechStackCategories :many
SELECT * FROM tech_stack_categories
ORDER BY id;

-- name: GetTechStackCategory :one
SELECT * FROM tech_stack_categories
WHERE id = $1 LIMIT 1;

-- name: UpdateTechStackCategory :one
UPDATE tech_stack_categories
SET name = $2,
    slug = $3,
    color = $4,
    updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteTechStackCategory :one
DELETE FROM tech_stack_categories
WHERE id = $1 RETURNING *;
