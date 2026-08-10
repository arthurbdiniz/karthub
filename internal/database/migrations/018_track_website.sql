-- +goose Up
ALTER TABLE tracks ADD COLUMN website TEXT;

-- +goose Down
ALTER TABLE tracks DROP COLUMN website;
