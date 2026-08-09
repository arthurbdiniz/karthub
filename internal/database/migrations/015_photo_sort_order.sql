-- +goose Up
ALTER TABLE event_photos ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE event_photos DROP COLUMN sort_order;
