-- name: CreateSubject :one
INSERT INTO subject.subjects (code, name, credits, description, department_id, weekly_hours, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetSubjectByID :one
SELECT * FROM subject.subjects WHERE id = $1;

-- name: GetSubjectByCode :one
SELECT * FROM subject.subjects WHERE code = $1;

-- name: ListSubjects :many
SELECT * FROM subject.subjects ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListSubjectsByDepartment :many
SELECT * FROM subject.subjects WHERE department_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountSubjects :one
SELECT COUNT(*) FROM subject.subjects;

-- name: CountSubjectsByDepartment :one
SELECT COUNT(*) FROM subject.subjects WHERE department_id = $1;

-- name: UpdateSubject :one
UPDATE subject.subjects
SET code         = COALESCE(sqlc.narg('code'), code),
    name         = COALESCE(sqlc.narg('name'), name),
    credits      = COALESCE(sqlc.narg('credits'), credits),
    description  = COALESCE(sqlc.narg('description'), description),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    weekly_hours = COALESCE(sqlc.narg('weekly_hours'), weekly_hours),
    is_active    = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at   = NOW()
WHERE id = sqlc.arg('id') RETURNING *;

-- name: DeleteSubject :exec
DELETE FROM subject.subjects WHERE id = $1;

-- name: ListAllSubjectIDs :many
SELECT id FROM subject.subjects ORDER BY id;
