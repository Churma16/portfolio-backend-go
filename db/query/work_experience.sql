-- name: CreateWorkExperience :one
INSERT INTO work_experiences (company, position, location, start_date, end_date, is_current, description)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetWorkExperiences :one
SELECT *
FROM work_experiences
WHERE id = $1 LIMIT 1;

-- name: ListWorkExperiences :many
SELECT *
FROM work_experiences
ORDER BY created_at DESC;

-- name: AttachTechStackToWorkExperience :exec
INSERT INTO work_experience_tech_stacks (work_experience_id, tech_stack_id)
VALUES ($1, $2);

-- name: AttachTagToWorkExperience :exec
INSERT INTO work_experience_tags (work_experience_id, tag_id)
VALUES ($1, $2);

