-- +goose Up
CREATE SCHEMA IF NOT EXISTS timetable;

-- +goose Down
-- Schema drop is intentionally omitted (would require CASCADE)
