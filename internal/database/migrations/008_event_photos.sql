-- +goose Up
CREATE TABLE event_photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id),
    driver_id INTEGER NOT NULL REFERENCES drivers(id),
    filename TEXT NOT NULL,
    original_name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_event_photos_event ON event_photos(event_id);

-- +goose Down
DROP TABLE IF EXISTS event_photos;
