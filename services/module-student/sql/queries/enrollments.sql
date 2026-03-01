-- name: CreateEnrollmentRequest :one
INSERT INTO student.enrollment_requests (
    id,
    student_id,
    semester_id,
    offered_subject_id,
    subject_id,
    request_note
) VALUES (
    sqlc.arg(id),
    sqlc.arg(student_id),
    sqlc.arg(semester_id),
    sqlc.arg(offered_subject_id),
    sqlc.arg(subject_id),
    sqlc.arg(request_note)
)
RETURNING *;

-- name: GetEnrollmentRequest :one
SELECT *
FROM student.enrollment_requests
WHERE id = $1;

-- name: ReviewEnrollmentRequest :one
UPDATE student.enrollment_requests
SET status = sqlc.arg(status),
    admin_note = sqlc.arg(admin_note),
    reviewed_by = sqlc.arg(reviewed_by),
    reviewed_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'pending'
RETURNING *;

-- name: ListEnrollmentRequests :many
SELECT *
FROM student.enrollment_requests
WHERE (sqlc.narg(student_id)::uuid IS NULL OR student_id = sqlc.narg(student_id)::uuid)
  AND (sqlc.narg(semester_id)::uuid IS NULL OR semester_id = sqlc.narg(semester_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY requested_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountEnrollmentRequests :one
SELECT COUNT(*)
FROM student.enrollment_requests
WHERE (sqlc.narg(student_id)::uuid IS NULL OR student_id = sqlc.narg(student_id)::uuid)
  AND (sqlc.narg(semester_id)::uuid IS NULL OR semester_id = sqlc.narg(semester_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text);

-- name: GetStudentEnrollments :many
SELECT *
FROM student.enrollment_requests
WHERE student_id = sqlc.arg(student_id)
  AND (sqlc.narg(semester_id)::uuid IS NULL OR semester_id = sqlc.narg(semester_id)::uuid)
ORDER BY requested_at DESC;

-- name: ListPassedEnrollmentSubjectIDs :many
SELECT subject_id
FROM student.enrollment_requests
WHERE student_id = $1
  AND status = 'completed'
ORDER BY requested_at DESC;

-- name: AppendEnrollmentEvent :exec
INSERT INTO student.event_store (aggregate_id, aggregate_type, event_type, payload)
VALUES ($1, 'enrollment_request', $2, $3);
