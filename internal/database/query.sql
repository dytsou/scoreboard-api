-- name: Get :one
SELECT * FROM scoreboards
WHERE id = $1 LIMIT 1;

-- name: List :many
SELECT * FROM scoreboards
ORDER BY created_at DESC;

-- name: Create :one
INSERT INTO scoreboards (
    id, name, created_at, updated_at
) VALUES (
    gen_random_uuid(),
    $1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) RETURNING *;

-- name: Update :one
UPDATE scoreboards
SET name = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: Delete :exec
DELETE FROM scoreboards
WHERE id = $1;
