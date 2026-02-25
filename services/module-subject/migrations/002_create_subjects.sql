-- +goose Up
CREATE TABLE subject.subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    credits INT NOT NULL DEFAULT 3,
    description TEXT NOT NULL DEFAULT '',
    department_id VARCHAR(255) NOT NULL DEFAULT '',
    weekly_hours INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subjects_code ON subject.subjects(code);
CREATE INDEX idx_subjects_department ON subject.subjects(department_id);
CREATE INDEX idx_subjects_active ON subject.subjects(is_active);

-- +goose Down
DROP TABLE IF EXISTS subject.subjects;
