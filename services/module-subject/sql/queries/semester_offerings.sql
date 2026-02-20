-- name: CreateSemesterOffering :one
INSERT INTO subject.semester_offerings (subject_id, semester_id, max_enrollment)
VALUES ($1, $2, $3) RETURNING *;

-- name: ListSemesterOfferingsBySubject :many
SELECT * FROM subject.semester_offerings WHERE subject_id = $1 ORDER BY created_at DESC;

-- name: ListSemesterOfferingsBySemester :many
SELECT * FROM subject.semester_offerings WHERE semester_id = $1 ORDER BY created_at DESC;

-- name: GetSemesterOffering :one
SELECT * FROM subject.semester_offerings WHERE id = $1;

-- name: DeleteSemesterOffering :exec
DELETE FROM subject.semester_offerings WHERE id = $1;
