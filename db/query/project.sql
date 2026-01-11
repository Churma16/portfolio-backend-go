-- name: CreateProject :one
INSERT INTO projects (title, slug, thumbnail, content, category_id, is_featured)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetProject :one
SELECT *
FROM projects
WHERE id = $1 LIMIT 1;

-- name: ListProjects :many
SELECT *
FROM projects
ORDER BY created_at DESC;

-- name: AttachTechStackToProject :exec
INSERT INTO project_tech_stacks (project_id, tech_stack_id)
VALUES ($1, $2);

-- name: AttachTagToProject :exec
INSERT INTO project_tags (project_id, tag_id)
VALUES ($1, $2);

