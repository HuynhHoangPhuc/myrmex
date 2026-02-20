-- +goose Up
CREATE SCHEMA IF NOT EXISTS subject;

-- +goose Down
DROP SCHEMA IF EXISTS subject CASCADE;
