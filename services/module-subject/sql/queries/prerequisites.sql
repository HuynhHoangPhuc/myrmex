-- name: AddPrerequisite :one
INSERT INTO subject.prerequisites (subject_id, prerequisite_id, type, priority)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: RemovePrerequisite :exec
DELETE FROM subject.prerequisites WHERE subject_id = $1 AND prerequisite_id = $2;

-- name: ListPrerequisitesBySubject :many
SELECT * FROM subject.prerequisites WHERE subject_id = $1 ORDER BY priority ASC;

-- name: ListAllPrerequisites :many
SELECT * FROM subject.prerequisites ORDER BY subject_id, priority ASC;

-- name: GetPrerequisite :one
SELECT * FROM subject.prerequisites WHERE subject_id = $1 AND prerequisite_id = $2;
