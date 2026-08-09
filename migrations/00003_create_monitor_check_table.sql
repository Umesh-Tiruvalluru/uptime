-- +goose Up
CREATE TABLE IF NOT EXISTS monitor_checks (
    id UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    response_time INTEGER NOT NULL,
    status_code INTEGER NOT NULL,
    error_message TEXT NOT NULL,
    monitor_id UUID REFERENCES monitors(id) ON DELETE CASCADE,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS monitor_checks;
