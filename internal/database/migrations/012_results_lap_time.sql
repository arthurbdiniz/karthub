-- +goose Up
ALTER TABLE results ADD COLUMN best_lap_time TEXT;

-- +goose Down
ALTER TABLE results DROP COLUMN best_lap_time;
