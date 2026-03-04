-- +goose Up
ALTER TABLE core.users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE core.users ADD COLUMN IF NOT EXISTS oauth_subject VARCHAR(255);
ALTER TABLE core.users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- OAuth users have no password; allow empty string as placeholder
-- (password_hash stays non-null to avoid schema type changes)

-- Unique index: one account per provider+subject combo
CREATE UNIQUE INDEX idx_users_oauth ON core.users(oauth_provider, oauth_subject)
  WHERE oauth_provider IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS core.idx_users_oauth;
ALTER TABLE core.users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE core.users DROP COLUMN IF EXISTS oauth_subject;
ALTER TABLE core.users DROP COLUMN IF EXISTS oauth_provider;
