CREATE TABLE IF NOT EXISTS tournaments (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    password_hash TEXT,
    max_players INTEGER NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ(0) NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);
