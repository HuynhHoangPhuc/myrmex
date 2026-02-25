-- +goose Up
-- Conversations table: stores per-user AI chat sessions with full message history as JSONB.
-- Chosen over separate messages table for MVP simplicity; JSONB is efficient for append-only
-- sequential reads of the full conversation context window.
CREATE TABLE IF NOT EXISTS core.conversations (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    title      VARCHAR(200),
    messages   JSONB       NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON core.conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON core.conversations(updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS core.idx_conversations_updated_at;
DROP INDEX IF EXISTS core.idx_conversations_user_id;
DROP TABLE IF EXISTS core.conversations;
