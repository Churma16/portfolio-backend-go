-- name: CreateMessage :one
INSERT INTO messages (name, email, content)
VALUES ($1, $2, $3) RETURNING *;