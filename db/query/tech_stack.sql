Insert into tech_stacks (name,slug,icon)
values ($1,$2,$3)
returning *;

-- name: ListTechStacks :many
select *
from tech_stacks;

