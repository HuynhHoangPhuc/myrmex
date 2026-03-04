-- +goose Up
CREATE TABLE notification.email_queue (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID        NOT NULL REFERENCES notification.notifications(id),
    recipient_email VARCHAR(255) NOT NULL,
    subject         VARCHAR(255) NOT NULL,
    html_body       TEXT        NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|sent|failed
    retry_count     INT         NOT NULL DEFAULT 0,
    max_retries     INT         NOT NULL DEFAULT 3,
    next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ
);

CREATE INDEX idx_email_queue_pending ON notification.email_queue (next_retry_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS notification.email_queue;
