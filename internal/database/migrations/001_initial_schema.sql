-- +goose Up

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE drivers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id),
    name TEXT NOT NULL,
    nickname TEXT,
    avatar TEXT,
    number INTEGER,
    bio TEXT,
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    country TEXT NOT NULL,
    city TEXT NOT NULL,
    length_meters INTEGER,
    indoor BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE championships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    season TEXT NOT NULL,
    points_system TEXT NOT NULL DEFAULT 'f1',
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    championship_id INTEGER REFERENCES championships(id),
    track_id INTEGER NOT NULL REFERENCES tracks(id),
    date DATE NOT NULL,
    time TIME,
    max_drivers INTEGER NOT NULL DEFAULT 20,
    entry_fee_cents INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id),
    driver_id INTEGER NOT NULL REFERENCES drivers(id),
    status TEXT NOT NULL DEFAULT 'confirmed',
    position INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, driver_id)
);

CREATE TABLE results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id),
    driver_id INTEGER NOT NULL REFERENCES drivers(id),
    position INTEGER NOT NULL,
    fastest_lap BOOLEAN NOT NULL DEFAULT 0,
    dnf BOOLEAN NOT NULL DEFAULT 0,
    penalty_seconds INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, driver_id)
);

CREATE TABLE championship_points (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    championship_id INTEGER NOT NULL REFERENCES championships(id),
    event_id INTEGER NOT NULL REFERENCES events(id),
    driver_id INTEGER NOT NULL REFERENCES drivers(id),
    points INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(championship_id, event_id, driver_id)
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    data TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_events_date ON events(date);
CREATE INDEX idx_events_championship ON events(championship_id);
CREATE INDEX idx_bookings_event ON bookings(event_id);
CREATE INDEX idx_bookings_driver ON bookings(driver_id);
CREATE INDEX idx_results_event ON results(event_id);
CREATE INDEX idx_results_driver ON results(driver_id);
CREATE INDEX idx_championship_points_championship ON championship_points(championship_id);
CREATE INDEX idx_championship_points_driver ON championship_points(driver_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- +goose Down

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS championship_points;
DROP TABLE IF EXISTS results;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS championships;
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS drivers;
DROP TABLE IF EXISTS users;
