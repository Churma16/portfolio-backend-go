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
SET name             = $2,
    headline         = $3,
    role             = $4,
    bio_short        = $5,
    bio_long         = $6,
    location         = $7,
    is_hireable      = $8,
    avatar           = $9,
    cv_files         = $10,
    hero_image_codes = $11,
    socials          = $12,
    updated_at       = now()
WHERE user_id = $1 RETURNING *;