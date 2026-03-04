-- +goose Up
CREATE TABLE notification.notifications (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    type       VARCHAR(50) NOT NULL,
    channel    VARCHAR(20) NOT NULL DEFAULT 'in_app', -- 'email' | 'in_app'
    title      VARCHAR(255) NOT NULL,
    body       TEXT        NOT NULL,
    metadata   JSONB,
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notif_user_unread ON notification.notifications (user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX idx_notif_user ON notification.notifications (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS notification.notifications;
