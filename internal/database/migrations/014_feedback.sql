-- +goose Up
CREATE TABLE feedback (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    driver_id INTEGER NOT NULL REFERENCES drivers(id),
    event_id INTEGER REFERENCES events(id),
    message TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_feedback_event ON feedback(event_id);

-- +goose Down
DROP TABLE IF EXISTS feedback;
