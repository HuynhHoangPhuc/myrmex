-- +goose Up
CREATE TABLE core.notifications (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    type       VARCHAR(50) NOT NULL, -- "schedule_published", "enrollment_approved", etc.
    title      VARCHAR(255) NOT NULL,
    body       TEXT        NOT NULL,
    data       JSONB,               -- { resource_type, resource_id, link }
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Efficient unread badge count + list queries
CREATE INDEX idx_notifications_user_unread ON core.notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;
CREATE INDEX idx_notifications_user ON core.notifications (user_id, created_at DESC);

CREATE TABLE core.notification_preferences (
    user_id        UUID    NOT NULL PRIMARY KEY REFERENCES core.users(id) ON DELETE CASCADE,
    email_enabled  BOOLEAN NOT NULL DEFAULT true,
    inapp_enabled  BOOLEAN NOT NULL DEFAULT true,
    disabled_types TEXT[]  NOT NULL DEFAULT '{}', -- e.g. ["enrollment_requested"]
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS core.notification_preferences;
DROP TABLE IF EXISTS core.notifications;
