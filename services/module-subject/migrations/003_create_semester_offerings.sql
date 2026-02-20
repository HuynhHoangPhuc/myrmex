-- +goose Up
CREATE TABLE subject.semester_offerings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id UUID NOT NULL REFERENCES subject.subjects(id) ON DELETE CASCADE,
    semester_id UUID NOT NULL,
    max_enrollment INT NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_semester_offerings_subject ON subject.semester_offerings(subject_id);
CREATE INDEX idx_semester_offerings_semester ON subject.semester_offerings(semester_id);

-- +goose Down
DROP TABLE IF EXISTS subject.semester_offerings;
