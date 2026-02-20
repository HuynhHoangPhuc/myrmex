-- +goose Up
CREATE TABLE timetable.semesters (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    year          INT NOT NULL,
    term          INT NOT NULL,
    start_date    DATE NOT NULL,
    end_date      DATE NOT NULL,
    offered_subject_ids UUID[] NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS timetable.semesters;
