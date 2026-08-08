-- +goose Up

-- Remove password requirement from users
-- SQLite doesn't support ALTER COLUMN, so we recreate the table
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users_new (id, email, name, role, created_at, updated_at)
SELECT id, email, username, role, created_at, updated_at FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

-- Magic link tokens
CREATE TABLE magic_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    used BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_magic_tokens_token ON magic_tokens(token);
CREATE INDEX idx_magic_tokens_email ON magic_tokens(email);

-- +goose Down

CREATE TABLE users_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users_old (id, username, email, password_hash, role, created_at, updated_at)
SELECT id, name, email, '', role, created_at, updated_at FROM users;

DROP TABLE users;
ALTER TABLE users_old RENAME TO users;

DROP TABLE IF EXISTS magic_tokens;
