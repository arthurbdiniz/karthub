-- +goose Up

ALTER TABLE drivers DROP COLUMN number;
-- avatar column already exists, we'll store base64 there

-- +goose Down

ALTER TABLE drivers ADD COLUMN number INTEGER;
