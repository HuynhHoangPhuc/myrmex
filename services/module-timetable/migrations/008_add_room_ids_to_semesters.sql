-- +goose Up
ALTER TABLE timetable.semesters ADD COLUMN room_ids UUID[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE timetable.semesters DROP COLUMN room_ids;
