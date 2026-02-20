-- +goose Up
CREATE TABLE timetable.time_slots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    semester_id UUID NOT NULL REFERENCES timetable.semesters(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_period INT NOT NULL CHECK (start_period >= 1),
    end_period   INT NOT NULL CHECK (end_period >= 1),
    CONSTRAINT valid_slot_period CHECK (start_period < end_period)
);

CREATE INDEX idx_time_slots_semester ON timetable.time_slots(semester_id);

-- +goose Down
DROP TABLE IF EXISTS timetable.time_slots;
