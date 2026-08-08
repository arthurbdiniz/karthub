-- +goose Up
ALTER TABLE events ADD COLUMN duration_minutes INTEGER NOT NULL DEFAULT 60;

-- +goose Down
ALTER TABLE events DROP COLUMN duration_minutes;
