-- name: CreateDepartment :one
INSERT INTO hr.departments (name, code)
VALUES ($1, $2) RETURNING *;

-- name: GetDepartmentByID :one
SELECT * FROM hr.departments WHERE id = $1;

-- name: ListDepartments :many
SELECT * FROM hr.departments ORDER BY name LIMIT $1 OFFSET $2;

-- name: CountDepartments :one
SELECT COUNT(*) FROM hr.departments;
