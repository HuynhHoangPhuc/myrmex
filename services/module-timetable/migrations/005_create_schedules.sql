-- +goose Up
CREATE TABLE timetable.schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    semester_id     UUID NOT NULL REFERENCES timetable.semesters(id),
    name            VARCHAR(200) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    score           FLOAT,
    hard_violations INT NOT NULL DEFAULT 0,
    soft_penalty    FLOAT NOT NULL DEFAULT 0,
    generated_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE timetable.schedule_entries (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id        UUID NOT NULL REFERENCES timetable.schedules(id) ON DELETE CASCADE,
    subject_id         UUID NOT NULL,
    teacher_id         UUID NOT NULL,
    room_id            UUID NOT NULL REFERENCES timetable.rooms(id),
    time_slot_id       UUID NOT NULL REFERENCES timetable.time_slots(id),
    is_manual_override BOOLEAN NOT NULL DEFAULT false,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, room_id, time_slot_id),
    UNIQUE (schedule_id, teacher_id, time_slot_id)
);

CREATE INDEX idx_schedule_entries_schedule  ON timetable.schedule_entries(schedule_id);
CREATE INDEX idx_schedule_entries_teacher   ON timetable.schedule_entries(teacher_id);
CREATE INDEX idx_schedule_entries_subject   ON timetable.schedule_entries(subject_id);

-- +goose Down
DROP TABLE IF EXISTS timetable.schedule_entries;
DROP TABLE IF EXISTS timetable.schedules;
