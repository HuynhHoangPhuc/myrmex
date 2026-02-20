-- name: CreateUser :one
INSERT INTO core.users (email, password_hash, full_name, role)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM core.users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM core.users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM core.users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM core.users;

-- name: UpdateUser :one
UPDATE core.users
SET full_name = COALESCE(sqlc.narg('full_name'), full_name),
    email = COALESCE(sqlc.narg('email'), email),
    role = COALESCE(sqlc.narg('role'), role),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id') RETURNING *;

-- name: DeleteUser :exec
DELETE FROM core.users WHERE id = $1;
