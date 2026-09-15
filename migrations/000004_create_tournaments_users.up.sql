CREATE TABLE IF NOT EXISTS tournaments_users (
    tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    PRIMARY KEY (tournament_id, user_id)
);
