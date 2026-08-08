-- +goose Up
ALTER TABLE events ADD COLUMN payment_info TEXT;

-- +goose Down
ALTER TABLE events DROP COLUMN payment_info;
