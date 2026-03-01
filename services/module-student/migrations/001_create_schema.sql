-- +goose Up
CREATE SCHEMA IF NOT EXISTS student;

-- +goose Down
DROP SCHEMA IF EXISTS student CASCADE;
