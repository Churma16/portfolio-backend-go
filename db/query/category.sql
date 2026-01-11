insert into categories (name, slug, icon)
values ($1, $2, $3)
returning *;

-- name: GetCategory :one
select *
from categories
where id = $1 limit 1;

-- name: ListCategories :many
select *
from categories
order by name;

