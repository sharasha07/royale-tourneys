package data

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/sharasha07/royale-tourneys/internal/validator"
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

func ValidateUser(v *validator.Validator, username, password *string) {
	if username != nil {
		v.Check(*username != "", "username", "must be provided")
		v.Check(utf8.RuneCountInString(*username) <= 10, "username", "must not have more than 10 characters")
	}

	if password != nil {
		v.Check(*password != "", "password", "must be provided")
		v.Check(utf8.RuneCountInString(*password) > 5, "password", "must be more than 5 characters")
		v.Check(utf8.RuneCountInString(*password) <= 15, "password", "must not have more than 15 characters")
	}
}

func ValidateGameTag(v *validator.Validator, gameTag, token string) error {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	endpoint := fmt.Sprintf("https://api.clashroyale.com/v1/players/%s", url.PathEscape(gameTag))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error while making request to clash API with status code: %d", resp.StatusCode)
	}

	v.Check(resp.StatusCode != http.StatusNotFound, "game_tag", "invalid")

	return nil
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

func (m DBModel) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET username = $1, password_hash = $2, game_tag = $3, version = version + 1
		WHERE id = $4 and version = $5
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []any{user.Username, user.PasswordHash, user.GameTag, user.ID, user.Version}

	err := m.pool.QueryRow(ctx, query, args...).Scan(
		&user.Version,
	)

	return err
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
