-- name: CreateSemester :one
INSERT INTO timetable.semesters (name, year, term, start_date, end_date, offered_subject_ids)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSemesterByID :one
SELECT * FROM timetable.semesters WHERE id = $1;

-- name: ListSemesters :many
SELECT * FROM timetable.semesters ORDER BY year DESC, term DESC LIMIT $1 OFFSET $2;

-- name: CountSemesters :one
SELECT COUNT(*) FROM timetable.semesters;

-- name: AddOfferedSubject :one
UPDATE timetable.semesters
SET offered_subject_ids = array_append(offered_subject_ids, $2)
WHERE id = $1
RETURNING *;

-- name: RemoveOfferedSubject :one
UPDATE timetable.semesters
SET offered_subject_ids = array_remove(offered_subject_ids, $2)
WHERE id = $1
RETURNING *;

-- name: CreateTimeSlot :one
INSERT INTO timetable.time_slots (semester_id, day_of_week, start_period, end_period)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTimeSlotsBySemester :many
SELECT * FROM timetable.time_slots WHERE semester_id = $1 ORDER BY day_of_week, start_period;
