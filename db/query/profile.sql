-- name: CreateProfile :one
INSERT INTO profiles (user_id, name, headline, role, bio_short, bio_long, location,
                      is_hireable, avatar, cv_files, hero_image_codes, socials)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *;

-- name: GetProfileByUserId :one
SELECT *
FROM profiles
WHERE user_id = $1 LIMIT 1;

-- name: UpdateProfile :one
UPDATE profiles
SET name      = $2,
    headline  = $3,
    bio_short = $4
WHERE user_id = $1 RETURNING *;