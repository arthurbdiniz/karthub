-- +goose Up
ALTER TABLE drivers ADD COLUMN country_code TEXT;

-- +goose Down
ALTER TABLE drivers DROP COLUMN country_code;
