-- +goose Up
CREATE TABLE hr.teacher_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES hr.teachers(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_period INT NOT NULL,
    end_period INT NOT NULL,
    CONSTRAINT valid_period_range CHECK (start_period < end_period),
    UNIQUE(teacher_id, day_of_week, start_period)
);

CREATE INDEX idx_availability_teacher ON hr.teacher_availability(teacher_id);

-- +goose Down
DROP TABLE IF EXISTS hr.teacher_availability;
