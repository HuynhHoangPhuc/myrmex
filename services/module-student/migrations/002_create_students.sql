-- +goose Up
CREATE TABLE IF NOT EXISTS student.students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_code VARCHAR(20) UNIQUE NOT NULL,
    user_id UUID,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    department_id UUID NOT NULL,
    enrollment_year INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_students_dept ON student.students(department_id);
CREATE INDEX IF NOT EXISTS idx_students_code ON student.students(student_code);
CREATE INDEX IF NOT EXISTS idx_students_email ON student.students(email);
CREATE INDEX IF NOT EXISTS idx_students_status ON student.students(status) WHERE is_active = true;

-- +goose Down
DROP TABLE IF EXISTS student.students;
