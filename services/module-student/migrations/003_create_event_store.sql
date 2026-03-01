-- +goose Up
CREATE TABLE IF NOT EXISTS student.event_store (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    version BIGINT NOT NULL DEFAULT 1,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_student_events_aggregate ON student.event_store(aggregate_id, aggregate_type);

-- +goose Down
DROP TABLE IF EXISTS student.event_store;
