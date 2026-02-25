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
    (schedule_id, subject_id, teacher_id, room_id, time_slot_id, is_manual_override,
     subject_name, subject_code, teacher_name, department_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetScheduleEntry :one
SELECT * FROM timetable.schedule_entries WHERE id = $1;

-- name: ListEntriesBySchedule :many
SELECT
    e.*,
    ts.day_of_week, ts.start_period, ts.end_period,
    r.name AS room_name
FROM timetable.schedule_entries e
JOIN timetable.time_slots ts ON e.time_slot_id = ts.id
JOIN timetable.rooms r       ON e.room_id      = r.id
WHERE e.schedule_id = $1
ORDER BY ts.day_of_week, ts.start_period;

-- name: ListSchedulesPaged :many
SELECT * FROM timetable.schedules
WHERE ($1::uuid IS NULL OR NOT $1::uuid = '00000000-0000-0000-0000-000000000000'::uuid AND semester_id = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSchedules :one
SELECT COUNT(*) FROM timetable.schedules
WHERE ($1::uuid IS NULL OR NOT $1::uuid = '00000000-0000-0000-0000-000000000000'::uuid AND semester_id = $1);

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
