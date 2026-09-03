package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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

func (m DBModel) GetUserByID(id int) (User, error) {
	query := `
		SELECT id, username, password_hash, game_tag, profile_picture, created_at, version
		FROM users
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u User
	err := m.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
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

func (m DBModel) DeleteUser(id int) error {
	query := `
		DELETE FROM users
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := m.pool.Exec(ctx, query, id)
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return err
}
