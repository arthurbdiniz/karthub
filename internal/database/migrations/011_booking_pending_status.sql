-- +goose Up
-- Change default booking status from 'confirmed' to 'pending'
-- Existing confirmed bookings stay as-is

-- +goose Down
