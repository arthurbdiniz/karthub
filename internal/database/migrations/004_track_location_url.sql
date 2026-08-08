-- +goose Up
ALTER TABLE tracks ADD COLUMN location_url TEXT;

-- +goose Down
ALTER TABLE tracks DROP COLUMN location_url;
