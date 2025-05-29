-- name: GetByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ExistsByEmail :one
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
) AS email_exists;

-- name: Create :one
INSERT INTO users (
    id, email, name, given_name, family_name, picture, email_verified, locale, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    $1, $2, $3, $4, $5, $6, $7,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
RETURNING *;

-- name: Update :one
UPDATE users SET 
    name = $2,
    given_name = $3,
    family_name = $4,
    picture = $5,
    email_verified = $6,
    locale = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE email = $1
RETURNING *;

-- name: Delete :exec
DELETE FROM users WHERE id = $1;