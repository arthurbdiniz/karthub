-- +goose Up
ALTER TABLE tracks ADD COLUMN map_embed TEXT;

-- +goose Down
ALTER TABLE tracks DROP COLUMN map_embed;
