-- +goose Up
CREATE TABLE users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    username      TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL,
    avatar_key    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
