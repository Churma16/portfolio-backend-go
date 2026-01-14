-- name: CreateProject :one
INSERT INTO projects (title, slug, thumbnail, content, demo_url, repo_url, is_featured, published_at, category_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: GetProject :one
SELECT *
FROM projects
WHERE id = $1 LIMIT 1;

-- name: GetProjects :many
SELECT *
FROM projects
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET title       = $2,
    slug        = $3,
    thumbnail   = $4,
    content     = $5,
    demo_url    = $6,
    repo_url    = $7,
    category_id = $8,
    updated_at  = now()
WHERE id = $1 RETURNING *;

-- name: DeleteProject :exec
DELETE
FROM projects
WHERE id = $1;

-- =============================================
-- PIVOT TABLE QUERIES
-- =============================================

-- name: AttachTechStackToProject :exec
INSERT INTO project_tech_stacks (project_id, tech_stack_id)
VALUES ($1, $2);

-- name: AttachTagToProject :exec
INSERT INTO project_tags (project_id, tag_id)
VALUES ($1, $2);

-- name: AddTechStackToProject :exec
INSERT INTO project_tech_stacks (project_id, tech_stack_id)
VALUES ($1, $2);

-- name: AddTagToProject :exec
INSERT INTO project_tags (project_id, tag_id)
VALUES ($1, $2);

-- name: DeleteProjectTechStacks :exec
DELETE
FROM project_tech_stacks
WHERE project_id = $1;

-- name: DeleteProjectTags :exec
DELETE
FROM project_tags
WHERE project_id = $1;

-- name: GetTechStacksByProjectID :many
SELECT ts.*
FROM tech_stacks ts
         JOIN project_tech_stacks pts ON ts.id = pts.tech_stack_id
WHERE pts.project_id = $1;

-- name: GetTagsByProjectID :many
SELECT t.*
FROM tags t
         JOIN project_tags pt ON t.id = pt.tag_id
WHERE pt.project_id = $1;
