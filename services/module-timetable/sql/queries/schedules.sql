-- name: CreateSchedule :one
INSERT INTO timetable.schedules (semester_id, name, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetScheduleByID :one
SELECT * FROM timetable.schedules WHERE id = $1;

-- name: ListSchedulesBySemester :many
SELECT * FROM timetable.schedules WHERE semester_id = $1 ORDER BY created_at DESC;

-- name: UpdateScheduleResult :one
UPDATE timetable.schedules
SET status = $2, score = $3, hard_violations = $4, soft_penalty = $5, generated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateScheduleStatus :one
UPDATE timetable.schedules SET status = $2 WHERE id = $1 RETURNING *;

-- name: CreateScheduleEntry :one
INSERT INTO timetable.schedule_entries
    (schedule_id, subject_id, teacher_id, room_id, time_slot_id, is_manual_override)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetScheduleEntry :one
SELECT * FROM timetable.schedule_entries WHERE id = $1;

-- name: ListEntriesBySchedule :many
SELECT * FROM timetable.schedule_entries WHERE schedule_id = $1 ORDER BY created_at;

-- name: UpdateScheduleEntry :one
UPDATE timetable.schedule_entries
SET teacher_id = $2, room_id = $3, time_slot_id = $4, is_manual_override = $5
WHERE id = $1
RETURNING *;

-- name: DeleteScheduleEntry :exec
DELETE FROM timetable.schedule_entries WHERE id = $1;

-- name: DeleteEntriesBySchedule :exec
DELETE FROM timetable.schedule_entries WHERE schedule_id = $1;

-- name: AppendEvent :exec
INSERT INTO timetable.event_store (aggregate_id, aggregate_type, event_type, payload)
VALUES ($1, $2, $3, $4);
