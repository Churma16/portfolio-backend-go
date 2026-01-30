-- name: CreateUser :one
INSERT INTO users (email, password)
VALUES ($1, $2)
    RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateUserPassword :one
UPDATE users
SET password = $2
WHERE id = $1 RETURNING *;