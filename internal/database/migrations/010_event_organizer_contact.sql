-- +goose Up
ALTER TABLE events ADD COLUMN organizer_contact TEXT;

-- +goose Down
ALTER TABLE events DROP COLUMN organizer_contact;
