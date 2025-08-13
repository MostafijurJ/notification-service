package db

import (
	"context"
	"database/sql"
	"fmt"
)

// RunMigrations applies minimal schema to support notifications, preferences, DND, groups, campaigns, and delivery attempts.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	sqlStmt := `
CREATE TABLE IF NOT EXISTS notification_types (
  id SERIAL PRIMARY KEY,
  key TEXT UNIQUE NOT NULL,
  category TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS templates (
  id BIGSERIAL PRIMARY KEY,
  type_key TEXT REFERENCES notification_types(key) ON DELETE CASCADE,
  channel TEXT NOT NULL,
  locale TEXT NOT NULL DEFAULT 'en',
  subject TEXT,
  body TEXT NOT NULL,
  version INT NOT NULL DEFAULT 1,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (type_key, channel, locale, version)
);

CREATE TABLE IF NOT EXISTS user_channel_preferences (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type_key TEXT NOT NULL REFERENCES notification_types(key),
  channel TEXT NOT NULL,
  opted_in BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, type_key, channel)
);

CREATE TABLE IF NOT EXISTS user_dnd_windows (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL UNIQUE,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  timezone TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS groups (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

CREATE TABLE IF NOT EXISTS group_members (
  group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL,
  PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS campaigns (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  type_key TEXT NOT NULL REFERENCES notification_types(key),
  channel TEXT NOT NULL,
  segment_group_id BIGINT REFERENCES groups(id),
  scheduled_at TIMESTAMPTZ,
  priority TEXT NOT NULL DEFAULT 'low',
  status TEXT NOT NULL DEFAULT 'scheduled',
  created_by TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
  id BIGSERIAL PRIMARY KEY,
  idempotency_key TEXT,
  user_id BIGINT,
  campaign_id BIGINT REFERENCES campaigns(id),
  type_key TEXT NOT NULL REFERENCES notification_types(key),
  channel TEXT NOT NULL,
  payload JSONB NOT NULL,
  priority TEXT NOT NULL,
  scheduled_at TIMESTAMPTZ,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);

CREATE TABLE IF NOT EXISTS delivery_attempts (
  id BIGSERIAL PRIMARY KEY,
  notification_id BIGINT REFERENCES notifications(id) ON DELETE CASCADE,
  attempt_no INT NOT NULL,
  provider_message_id TEXT,
  status TEXT NOT NULL,
  error_code TEXT,
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS device_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  platform TEXT NOT NULL,
  token TEXT NOT NULL,
  provider TEXT NOT NULL DEFAULT 'fcm',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  last_seen_at TIMESTAMPTZ,
  UNIQUE (provider, token)
);
CREATE INDEX IF NOT EXISTS idx_device_tokens_user ON device_tokens (user_id);

CREATE TABLE IF NOT EXISTS inapp_notifications (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  type_key TEXT NOT NULL,
  title TEXT,
  body TEXT NOT NULL,
  metadata JSONB,
  read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_inapp_user_read_created ON inapp_notifications (user_id, read, created_at DESC);
`

	if _, err := db.ExecContext(ctx, sqlStmt); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}