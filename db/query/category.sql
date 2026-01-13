-- name: CreateCategory :one
insert into categories (name, slug, color)
values ($1, $2, $3) returning *;

-- name: GetCategories :many
select *
from categories;

-- name: GetCategory :one
select *
from categories
where id = $1 limit 1;

-- name: UpdateCategory :one
update categories
set name       = $2,
    slug       = $3,
    color      = $4,
    updated_at = now()
where id = $1 returning *;

-- name: DeleteCategory :one
delete
from categories
where id = $1 returning *;
