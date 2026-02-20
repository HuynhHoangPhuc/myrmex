-- +goose Up
CREATE TABLE core.module_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    version VARCHAR(20) NOT NULL,
    grpc_address VARCHAR(255) NOT NULL,
    health_status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_health_check TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS core.module_registry;
