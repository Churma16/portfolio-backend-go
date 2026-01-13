-- name: CreateTechStack :one
Insert into tech_stacks (name, slug, icon)
values ($1, $2, $3) returning *;

-- name: GetTechStacks :many
select *
from tech_stacks
order by column_order;

-- name: GetTechStack :one
select *
from tech_stacks
where id = $1 limit 1;

-- name: UpdateTechStack :one
update tech_stacks
set name       = $2,
    slug       = $3,
    icon       = $4,
    updated_at = now()
where id = $1 returning *;

-- name: DeleteTechStack :one
delete
from tech_stacks
where id = $1 returning *;
