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

-- name: UpdateWorkExperience :one
update work_experiences
set company     = $2,
    position    = $3,
    location    = $4,
    start_date  = $5,
    end_date    = $6,
    is_current  = $7,
    description = $8,
    updated_at  = now()
where id = $1 RETURNING *;

-- name: DeleteWorkExperience :one
DELETE
FROM work_experiences
WHERE id = $1 RETURNING *;
-- =============================================
-- PIVOT TABLE QUERIES
-- =============================================
    
-- name: AttachTechStackToWorkExperience :exec
INSERT INTO work_experience_tech_stacks (work_experience_id, tech_stack_id)
VALUES ($1, $2);

-- name: AttachTagToWorkExperience :exec
INSERT INTO work_experience_tags (work_experience_id, tag_id)
VALUES ($1, $2);

-- name: AddTechStackToWorkExperience :exec
INSERT INTO work_experience_tech_stacks (work_experience_id, tech_stack_id)
VALUES ($1, $2);

-- name: AddTagToWorkExperience :exec
INSERT INTO work_experience_tags (work_experience_id, tag_id)
VALUES ($1, $2);

-- name: DeleteWorkExperienceTechStacks :exec
DELETE
FROM work_experience_tech_stacks
WHERE work_experience_id = $1;

-- name: DeleteWorkExperienceTags :exec
DELETE
FROM work_experience_tags
WHERE work_experience_id = $1;

-- name: GetTechStacksByWorkExperienceID :many
SELECT work_experience_tech_stacks.work_experience_id, tech_stacks.*
FROM tech_stacks
         JOIN work_experience_tech_stacks ON tech_stacks.id = work_experience_tech_stacks.tech_stack_id
WHERE work_experience_tech_stacks.work_experience_id = ANY (@work_experience_ids::int[]);

-- name: GetTagsByWorkExperienceID :many
SELECT work_experience_tags.work_experience_id, tags.*
FROM tags
         JOIN work_experience_tags ON tags.id = work_experience_tags.tag_id
WHERE work_experience_tags.work_experience_id = ANY (@work_experience_ids::int[]);


