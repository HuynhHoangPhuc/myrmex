-- name: AssignGrade :one
INSERT INTO student.grades (
    id,
    enrollment_id,
    grade_numeric,
    graded_by,
    notes
) VALUES (
    sqlc.arg(id),
    sqlc.arg(enrollment_id),
    sqlc.arg(grade_numeric)::double precision,
    sqlc.arg(graded_by),
    sqlc.arg(notes)
)
RETURNING
    id,
    enrollment_id,
    grade_numeric::double precision AS grade_numeric,
    grade_letter,
    graded_by,
    graded_at,
    notes;

-- name: UpdateGrade :one
UPDATE student.grades
SET grade_numeric = sqlc.arg(grade_numeric)::double precision,
    graded_by = sqlc.arg(graded_by),
    notes = sqlc.arg(notes),
    graded_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING
    id,
    enrollment_id,
    grade_numeric::double precision AS grade_numeric,
    grade_letter,
    graded_by,
    graded_at,
    notes;

-- name: GetGrade :one
SELECT
    id,
    enrollment_id,
    grade_numeric::double precision AS grade_numeric,
    grade_letter,
    graded_by,
    graded_at,
    notes
FROM student.grades
WHERE id = sqlc.arg(id);

-- name: MarkEnrollmentCompleted :exec
UPDATE student.enrollment_requests
SET status = 'completed'
WHERE id = sqlc.arg(enrollment_id);

-- name: GetStudentTranscript :many
SELECT
    er.id AS enrollment_id,
    er.student_id,
    er.semester_id,
    er.subject_id,
    er.offered_subject_id,
    er.status,
    er.requested_at,
    COALESCE(g.grade_numeric::double precision, 0::double precision) AS grade_numeric,
    g.grade_letter,
    g.graded_at
FROM student.enrollment_requests er
LEFT JOIN student.grades g ON g.enrollment_id = er.id
WHERE er.student_id = sqlc.arg(student_id)
  AND er.status IN ('approved', 'completed')
ORDER BY er.requested_at DESC;
