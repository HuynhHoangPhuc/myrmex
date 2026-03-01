-- +goose Up
CREATE TABLE analytics.dim_student (
    student_id      UUID PRIMARY KEY,
    student_code    VARCHAR(20),
    full_name       VARCHAR(255),
    department_id   UUID,
    enrollment_year INT,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE analytics.fact_enrollment (
    enrollment_id UUID PRIMARY KEY,
    student_id    UUID REFERENCES analytics.dim_student(student_id),
    subject_id    UUID REFERENCES analytics.dim_subject(subject_id),
    semester_id   UUID REFERENCES analytics.dim_semester(semester_id),
    status        VARCHAR(20),
    grade_numeric NUMERIC(4,2),
    grade_letter  VARCHAR(2),
    enrolled_at   TIMESTAMPTZ,
    graded_at     TIMESTAMPTZ
);

CREATE INDEX idx_fact_enrollment_student  ON analytics.fact_enrollment(student_id);
CREATE INDEX idx_fact_enrollment_semester ON analytics.fact_enrollment(semester_id);
CREATE INDEX idx_fact_enrollment_subject  ON analytics.fact_enrollment(subject_id);

-- +goose Down
DROP TABLE IF EXISTS analytics.fact_enrollment;
DROP TABLE IF EXISTS analytics.dim_student;
