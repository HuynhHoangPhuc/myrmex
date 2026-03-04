-- +goose Up
-- RBAC Phase 1: Add department_id to users for dept-scoped access
ALTER TABLE core.users ADD COLUMN department_id UUID;

-- +goose Down
ALTER TABLE core.users DROP COLUMN IF EXISTS department_id;
