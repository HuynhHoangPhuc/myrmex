-- +goose Up
-- Partitioned audit log table (range partition by month for instant retention drops)
CREATE TABLE core.audit_logs (
    id           UUID         NOT NULL DEFAULT gen_random_uuid(),
    user_id      UUID         NOT NULL,
    user_role    VARCHAR(50)  NOT NULL,
    action       VARCHAR(100) NOT NULL,  -- e.g. "hr.teacher.create"
    resource_type VARCHAR(50) NOT NULL,  -- e.g. "teacher"
    resource_id  UUID,
    old_value    JSONB,
    new_value    JSONB,
    ip_address   INET,
    user_agent   TEXT,
    status_code  INT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Pre-create 12 monthly partitions (2026-03 through 2027-02)
CREATE TABLE core.audit_logs_2026_03 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE core.audit_logs_2026_04 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE core.audit_logs_2026_05 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE core.audit_logs_2026_06 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE core.audit_logs_2026_07 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE core.audit_logs_2026_08 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE core.audit_logs_2026_09 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE core.audit_logs_2026_10 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE core.audit_logs_2026_11 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE core.audit_logs_2026_12 PARTITION OF core.audit_logs FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE core.audit_logs_2027_01 PARTITION OF core.audit_logs FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE core.audit_logs_2027_02 PARTITION OF core.audit_logs FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

-- Append-optimized index for time-range queries
CREATE INDEX idx_audit_created_brin   ON core.audit_logs USING BRIN (created_at);
-- Admin filter queries
CREATE INDEX idx_audit_user_date      ON core.audit_logs (user_id, created_at DESC);
CREATE INDEX idx_audit_resource_date  ON core.audit_logs (resource_type, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS core.audit_logs CASCADE;
