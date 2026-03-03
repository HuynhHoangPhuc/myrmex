-- name: CreateInviteCode :one
INSERT INTO student.invite_codes (code_hash, student_id, expires_at, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindInviteCodeByHash :one
SELECT * FROM student.invite_codes WHERE code_hash = $1;

-- name: MarkInviteCodeUsed :one
UPDATE student.invite_codes SET used_at = NOW(), used_by_user_id = $2
WHERE id = $1 AND used_at IS NULL
RETURNING *;

-- name: InvalidateStudentCodes :exec
UPDATE student.invite_codes SET expires_at = NOW()
WHERE student_id = $1 AND used_at IS NULL AND expires_at > NOW();
