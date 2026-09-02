package data

import (
	"context"
	"time"
)

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   []byte    `json:"-"`
	GameTag        *string   `json:"game_tag"`
	ProfilePicture *string   `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
	Version        int       `json:"version"`
}

func (m DBModel) CreateUser(username string, passwordHash []byte) (User, error) {
	query := `
		INSERT INTO users(username, password_hash)
		VALUES($1, $2)
		RETURNING id, game_tag, profile_picture, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	args := []any{username, passwordHash}
	err := m.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.GameTag,
		&u.ProfilePicture,
		&u.CreatedAt,
		&u.Version,
	)
	if err != nil {
		return User{}, err
	}

	return u, nil
}
