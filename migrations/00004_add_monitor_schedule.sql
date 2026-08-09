-- +goose Up
ALTER TABLE monitors
    ADD COLUMN IF NOT EXISTS interval_seconds INTEGER NOT NULL DEFAULT 60,
    ADD COLUMN IF NOT EXISTS next_check_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS last_status TEXT NOT NULL DEFAULT 'pending';

CREATE INDEX IF NOT EXISTS idx_monitors_next_check_at ON monitors (next_check_at);

-- +goose Down
DROP INDEX IF EXISTS idx_monitors_next_check_at;
ALTER TABLE monitors
    DROP COLUMN IF EXISTS last_status,
    DROP COLUMN IF EXISTS last_checked_at,
    DROP COLUMN IF EXISTS next_check_at,
    DROP COLUMN IF EXISTS interval_seconds;
