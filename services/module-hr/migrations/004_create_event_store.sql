-- +goose Up
CREATE TABLE hr.events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    version INT NOT NULL,
    data JSONB NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(aggregate_id, version)
);

CREATE INDEX idx_hr_events_aggregate ON hr.events(aggregate_id, version);
CREATE INDEX idx_hr_events_type ON hr.events(event_type);

CREATE TABLE hr.snapshots (
    aggregate_id UUID PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    version INT NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS hr.snapshots;
DROP TABLE IF EXISTS hr.events;
