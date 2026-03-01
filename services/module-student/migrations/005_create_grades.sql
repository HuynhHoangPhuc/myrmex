-- +goose Up
CREATE TABLE IF NOT EXISTS student.grades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL UNIQUE REFERENCES student.enrollment_requests(id),
    grade_numeric NUMERIC(4,2) NOT NULL CHECK (grade_numeric >= 0 AND grade_numeric <= 10),
    grade_letter VARCHAR(2) GENERATED ALWAYS AS (
        CASE
            WHEN grade_numeric >= 8.5 THEN 'A'
            WHEN grade_numeric >= 7.0 THEN 'B'
            WHEN grade_numeric >= 5.5 THEN 'C'
            WHEN grade_numeric >= 4.0 THEN 'D'
            ELSE 'F'
        END
    ) STORED,
    graded_by UUID NOT NULL,
    graded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_grades_enrollment ON student.grades(enrollment_id);

-- +goose Down
DROP TABLE IF EXISTS student.grades;
