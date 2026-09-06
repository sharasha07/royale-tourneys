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

var AnonymousUser *User

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   []byte    `json:"-"`
	GameTag        *string   `json:"game_tag"`
	ProfilePicture *string   `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
	Version        int       `json:"version"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
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

func ValidateGameTag(v *validator.Validator, gameTag, token string, client *http.Client) error {
	if gameTag == "" {
		v.Add("game_tag", "must be provided")
		return nil
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
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	case http.StatusNotFound:
		v.Add("game_tag", "invalid")
		return nil

	default:
		return fmt.Errorf("error while making request to clash API with status code: %d", resp.StatusCode)
	}
}

func (m DBModel) CreateUser(ctx context.Context, username string, passwordHash []byte) (User, error) {
	query := `
		INSERT INTO users(username, password_hash)
		VALUES($1, $2)
		RETURNING id, game_tag, profile_picture, created_at, version`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

func (m DBModel) GetUserByID(ctx context.Context, id int) (User, error) {
	query := `
		SELECT id, username, password_hash, game_tag, profile_picture, created_at, version
		FROM users
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

func (m DBModel) GetUserByUsername(ctx context.Context, username string) (User, error) {
	query := `
		SELECT id, username, password_hash, game_tag, profile_picture, created_at, version
		FROM users
		WHERE username = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var u User
	err := m.pool.QueryRow(ctx, query, username).Scan(
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

func (m DBModel) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $1, password_hash = $2, game_tag = $3, profile_picture = $4, version = version + 1
		WHERE id = $5 and version = $6
		RETURNING version`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := []any{user.Username, user.PasswordHash, user.GameTag, user.ProfilePicture, user.ID, user.Version}

	err := m.pool.QueryRow(ctx, query, args...).Scan(
		&user.Version,
	)

	return err
}

func (m DBModel) DeleteUser(ctx context.Context, id int) error {
	query := `
		DELETE FROM users
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := m.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
