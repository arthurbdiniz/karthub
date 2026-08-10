-- +goose Up
CREATE TABLE polls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    allow_multiple BOOLEAN NOT NULL DEFAULT 0,
    closes_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE poll_options (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    poll_id INTEGER NOT NULL REFERENCES polls(id),
    track_id INTEGER REFERENCES tracks(id),
    label TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE poll_votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    poll_id INTEGER NOT NULL REFERENCES polls(id),
    option_id INTEGER NOT NULL REFERENCES poll_options(id),
    driver_id INTEGER NOT NULL REFERENCES drivers(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(poll_id, option_id, driver_id)
);

CREATE INDEX idx_poll_votes_poll ON poll_votes(poll_id);
CREATE INDEX idx_poll_votes_option ON poll_votes(option_id);

-- +goose Down
DROP TABLE IF EXISTS poll_votes;
DROP TABLE IF EXISTS poll_options;
DROP TABLE IF EXISTS polls;
