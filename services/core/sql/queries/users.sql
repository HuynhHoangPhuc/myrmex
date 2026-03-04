-- name: CreateUser :one
INSERT INTO core.users (email, password_hash, full_name, role)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, role, is_active, department_id, created_at, updated_at
FROM core.users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, role, is_active, department_id, created_at, updated_at
FROM core.users WHERE email = $1;

-- name: ListUsers :many
SELECT id, email, password_hash, full_name, role, is_active, department_id, created_at, updated_at
FROM core.users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM core.users;

-- name: UpdateUser :one
UPDATE core.users
SET full_name = COALESCE(sqlc.narg('full_name'), full_name),
    email = COALESCE(sqlc.narg('email'), email),
    role = COALESCE(sqlc.narg('role'), role),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    updated_at = NOW()
WHERE id = sqlc.arg('id') RETURNING *;

-- name: DeleteUser :exec
DELETE FROM core.users WHERE id = $1;

-- name: UpdateUserRole :one
-- Used by admin role management API to set role + optional department scope
UPDATE core.users
SET role = $1,
    department_id = $2,
    updated_at = NOW()
WHERE id = $3
RETURNING id, email, password_hash, full_name, role, is_active, department_id, created_at, updated_at;

-- name: GetTeacherIDByUserID :one
-- Cross-schema lookup: find teacher record linked to a core user (for JWT population)
SELECT id FROM hr.teachers WHERE user_id = $1 LIMIT 1;

-- name: GetUserByOAuth :one
-- Find a user by their OAuth provider + subject (stable provider-issued user ID)
SELECT id, email, password_hash, full_name, role, is_active, department_id,
       oauth_provider, oauth_subject, avatar_url, created_at, updated_at
FROM core.users WHERE oauth_provider = $1 AND oauth_subject = $2;

-- name: UpsertOAuthUser :one
-- Link OAuth identity to an existing account (matched by email) or create a new one.
-- On conflict: update OAuth link and avatar; leave role/department unchanged.
INSERT INTO core.users (email, password_hash, full_name, role, oauth_provider, oauth_subject, avatar_url)
VALUES ($1, '', $2, $3, $4, $5, $6)
ON CONFLICT (email) DO UPDATE SET
    oauth_provider = EXCLUDED.oauth_provider,
    oauth_subject  = EXCLUDED.oauth_subject,
    avatar_url     = COALESCE(EXCLUDED.avatar_url, core.users.avatar_url),
    updated_at     = NOW()
RETURNING id, email, password_hash, full_name, role, is_active, department_id,
          oauth_provider, oauth_subject, avatar_url, created_at, updated_at;
