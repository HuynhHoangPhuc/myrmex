-- +goose Up
CREATE SCHEMA IF NOT EXISTS analytics;

-- Dimension: teachers
CREATE TABLE analytics.dim_teacher (
    teacher_id      UUID PRIMARY KEY,
    full_name       VARCHAR(200) NOT NULL,
    department_id   UUID,
    department_name VARCHAR(200),
    specializations TEXT[]       NOT NULL DEFAULT '{}',
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Dimension: departments
CREATE TABLE analytics.dim_department (
    department_id UUID PRIMARY KEY,
    name          VARCHAR(200) NOT NULL,
    code          VARCHAR(50)  NOT NULL,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Dimension: subjects
CREATE TABLE analytics.dim_subject (
    subject_id    UUID PRIMARY KEY,
    name          VARCHAR(200) NOT NULL,
    code          VARCHAR(50)  NOT NULL,
    credits       INT          NOT NULL DEFAULT 0,
    department_id UUID,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Dimension: semesters
CREATE TABLE analytics.dim_semester (
    semester_id UUID PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    year        INT          NOT NULL,
    term        VARCHAR(50)  NOT NULL,
    start_date  DATE,
    end_date    DATE,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Fact: teacher workload per semester per subject
CREATE TABLE analytics.fact_workload (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id    UUID        NOT NULL REFERENCES analytics.dim_teacher(teacher_id) ON DELETE CASCADE,
    semester_id   UUID        NOT NULL,
    subject_id    UUID        NOT NULL,
    hours_per_week NUMERIC(4,1) NOT NULL DEFAULT 0,
    total_hours   NUMERIC(6,1) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (teacher_id, semester_id, subject_id)
);

CREATE INDEX idx_fact_workload_semester  ON analytics.fact_workload(semester_id);
CREATE INDEX idx_fact_workload_teacher   ON analytics.fact_workload(teacher_id);

-- Fact: schedule entries (denormalised from timetable module)
CREATE TABLE analytics.fact_schedule_entry (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID        NOT NULL,
    semester_id UUID        NOT NULL,
    teacher_id  UUID        NOT NULL,
    subject_id  UUID        NOT NULL,
    room_id     UUID        NOT NULL,
    day_of_week INT         NOT NULL,
    period      INT         NOT NULL,
    is_assigned BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, day_of_week, period, room_id)
);

CREATE INDEX idx_fact_schedule_semester  ON analytics.fact_schedule_entry(semester_id);
CREATE INDEX idx_fact_schedule_teacher   ON analytics.fact_schedule_entry(teacher_id);

-- +goose Down
DROP TABLE IF EXISTS analytics.fact_schedule_entry;
DROP TABLE IF EXISTS analytics.fact_workload;
DROP TABLE IF EXISTS analytics.dim_semester;
DROP TABLE IF EXISTS analytics.dim_subject;
DROP TABLE IF EXISTS analytics.dim_department;
DROP TABLE IF EXISTS analytics.dim_teacher;
DROP SCHEMA IF EXISTS analytics;
