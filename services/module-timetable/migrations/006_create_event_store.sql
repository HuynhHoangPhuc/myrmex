-- +goose Up
CREATE TABLE timetable.event_store (
    id             BIGSERIAL PRIMARY KEY,
    aggregate_id   UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    payload        JSONB NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_store_aggregate ON timetable.event_store(aggregate_id, aggregate_type);

-- +goose Down
DROP TABLE IF EXISTS timetable.event_store;
