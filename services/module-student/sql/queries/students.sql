-- name: CreateStudent :one
INSERT INTO student.students (
    student_code,
    user_id,
    full_name,
    email,
    department_id,
    enrollment_year,
    status,
    is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetStudentByID :one
SELECT * FROM student.students WHERE id = $1 AND is_active = true;

-- name: ListStudents :many
SELECT *
FROM student.students
WHERE is_active = true
  AND (sqlc.narg(department_id)::uuid IS NULL OR department_id = sqlc.narg(department_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY full_name
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountStudents :one
SELECT COUNT(*)
FROM student.students
WHERE is_active = true
  AND (sqlc.narg(department_id)::uuid IS NULL OR department_id = sqlc.narg(department_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text);

-- name: UpdateStudent :one
UPDATE student.students
SET user_id = $2,
    full_name = $3,
    email = $4,
    department_id = $5,
    enrollment_year = $6,
    status = $7,
    is_active = $8,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: DeleteStudent :one
UPDATE student.students
SET is_active = false,
    updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING id;
