-- +goose Up
ALTER TABLE events ADD COLUMN status TEXT NOT NULL DEFAULT 'open';

-- +goose Down
ALTER TABLE events DROP COLUMN status;
