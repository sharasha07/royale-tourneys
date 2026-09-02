CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username CITEXT NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    game_tag TEXT UNIQUE,
    profile_picture TEXT,
    created_at TIMESTAMPTZ(0) NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);
