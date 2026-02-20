-- name: CreateTeacher :one
INSERT INTO hr.teachers (employee_code, full_name, email, phone, department_id, max_hours_per_week, title)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetTeacherByID :one
SELECT * FROM hr.teachers WHERE id = $1;

-- name: ListTeachers :many
SELECT * FROM hr.teachers WHERE is_active = true ORDER BY full_name LIMIT $1 OFFSET $2;

-- name: ListTeachersByDepartment :many
SELECT * FROM hr.teachers WHERE department_id = $1 AND is_active = true ORDER BY full_name LIMIT $2 OFFSET $3;

-- name: CountTeachers :one
SELECT COUNT(*) FROM hr.teachers WHERE is_active = true;

-- name: UpdateTeacher :one
UPDATE hr.teachers
SET full_name = $2, email = $3, phone = $4, department_id = $5,
    max_hours_per_week = $6, title = $7, is_active = $8, updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: DeleteTeacher :exec
UPDATE hr.teachers SET is_active = false, updated_at = NOW() WHERE id = $1;

-- name: SearchTeachersByName :many
SELECT * FROM hr.teachers
WHERE full_name ILIKE '%' || $1 || '%' AND is_active = true
ORDER BY full_name LIMIT $2 OFFSET $3;

-- name: AddSpecialization :exec
INSERT INTO hr.teacher_specializations (teacher_id, specialization)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveSpecialization :exec
DELETE FROM hr.teacher_specializations WHERE teacher_id = $1 AND specialization = $2;

-- name: ListSpecializations :many
SELECT specialization FROM hr.teacher_specializations WHERE teacher_id = $1;

-- name: SearchTeachersBySpecialization :many
SELECT t.* FROM hr.teachers t
JOIN hr.teacher_specializations ts ON t.id = ts.teacher_id
WHERE ts.specialization = $1 AND t.is_active = true
ORDER BY t.full_name;
