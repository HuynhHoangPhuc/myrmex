-- +goose Up
CREATE TABLE hr.teachers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_code VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    department_id UUID REFERENCES hr.departments(id),
    max_hours_per_week INT NOT NULL DEFAULT 20,
    title VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hr.teacher_specializations (
    teacher_id UUID NOT NULL REFERENCES hr.teachers(id) ON DELETE CASCADE,
    specialization VARCHAR(100) NOT NULL,
    PRIMARY KEY (teacher_id, specialization)
);

CREATE INDEX idx_teachers_dept ON hr.teachers(department_id);
CREATE INDEX idx_teachers_name ON hr.teachers(full_name);
CREATE INDEX idx_teacher_specs ON hr.teacher_specializations(specialization);

-- +goose Down
DROP TABLE IF EXISTS hr.teacher_specializations;
DROP TABLE IF EXISTS hr.teachers;
