-- +goose Up
-- RBAC Phase 1: Link teacher records to core.users for JWT token population
ALTER TABLE hr.teachers ADD COLUMN user_id UUID;

CREATE INDEX idx_teachers_user_id ON hr.teachers(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_teachers_user_id;
ALTER TABLE hr.teachers DROP COLUMN IF EXISTS user_id;
