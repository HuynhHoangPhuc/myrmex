-- +goose Up
CREATE TABLE IF NOT EXISTS student.enrollment_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES student.students(id),
    semester_id UUID NOT NULL,
    offered_subject_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    request_note TEXT NOT NULL DEFAULT '',
    admin_note TEXT NOT NULL DEFAULT '',
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID,
    UNIQUE(student_id, semester_id, offered_subject_id)
);

CREATE INDEX IF NOT EXISTS idx_enrollment_requests_student ON student.enrollment_requests(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_requests_semester ON student.enrollment_requests(semester_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_requests_status ON student.enrollment_requests(status);
CREATE INDEX IF NOT EXISTS idx_enrollment_requests_subject ON student.enrollment_requests(subject_id);

-- +goose Down
DROP TABLE IF EXISTS student.enrollment_requests;
