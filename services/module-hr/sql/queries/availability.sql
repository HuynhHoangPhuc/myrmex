-- name: UpsertAvailabilitySlot :one
INSERT INTO hr.teacher_availability (teacher_id, day_of_week, start_period, end_period)
VALUES ($1, $2, $3, $4)
ON CONFLICT (teacher_id, day_of_week, start_period)
DO UPDATE SET end_period = EXCLUDED.end_period
RETURNING *;

-- name: DeleteTeacherAvailability :exec
DELETE FROM hr.teacher_availability WHERE teacher_id = $1;

-- name: ListTeacherAvailability :many
SELECT * FROM hr.teacher_availability
WHERE teacher_id = $1
ORDER BY day_of_week, start_period;
