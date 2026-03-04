-- +goose Up
CREATE TABLE notification.preferences (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    channel    VARCHAR(20) NOT NULL, -- 'email' | 'in_app'
    enabled    BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, event_type, channel)
);

CREATE INDEX idx_prefs_user ON notification.preferences (user_id);

-- +goose Down
DROP TABLE IF EXISTS notification.preferences;
