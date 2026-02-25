-- +goose Up
ALTER TABLE timetable.schedule_entries
  ADD COLUMN subject_name   TEXT NOT NULL DEFAULT '',
  ADD COLUMN subject_code   TEXT NOT NULL DEFAULT '',
  ADD COLUMN teacher_name   TEXT NOT NULL DEFAULT '',
  ADD COLUMN department_id  UUID;

-- +goose Down
ALTER TABLE timetable.schedule_entries
  DROP COLUMN subject_name,
  DROP COLUMN subject_code,
  DROP COLUMN teacher_name,
  DROP COLUMN department_id;
