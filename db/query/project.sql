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
ORDER BY column_order ASC;

-- name: GetProjectByColumnOrder :one
SELECT *
FROM projects
WHERE column_order = $1 LIMIT 1;

-- name: UpdateProject :one
UPDATE projects
SET title       = $2,
    slug        = $3,
    thumbnail   = $4,
    content     = $5,
    demo_url    = $6,
    repo_url    = $7,
    category_id = $8,
    published_at = $9,
    updated_at  = now()
WHERE id = $1 RETURNING *;

-- name: UpdateProjectColumnOrder :one
UPDATE projects
SET column_order = $2,
    updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteProject :one
DELETE
FROM projects
WHERE id = $1 RETURNING *;

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
SELECT project_tech_stacks.project_id, tech_stacks.*
FROM tech_stacks
         JOIN project_tech_stacks ON tech_stacks.id = project_tech_stacks.tech_stack_id
WHERE project_tech_stacks.project_id = ANY (@project_ids::int[])
ORDER BY tech_stacks.column_order ASC;



-- name: GetTagsByProjectID :many
SELECT project_tags.project_id, tags.*
FROM tags
         JOIN project_tags ON tags.id = project_tags.tag_id
WHERE project_tags.project_id = ANY (@project_ids::int[]);

-- name: GetCategoriesByIDs :many
SELECT *
FROM categories
WHERE id = ANY (@category_ids::int[]);