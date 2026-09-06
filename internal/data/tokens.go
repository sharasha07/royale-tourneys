package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/pascaldekloe/jwt"
)

func NewJWTToken(userID int, jwtSecret string, ttl time.Duration) (string, error) {
	claims := jwt.Claims{
		Subject:   strconv.FormatInt(int64(userID), 10),
		Issued:    jwt.NewNumericTime(time.Now()),
		NotBefore: jwt.NewNumericTime(time.Now()),
		Expires:   jwt.NewNumericTime(time.Now().Add(ttl)),
		Issuer:    "github.com/sharasha07/royale-tourneys",
		Audiences: []string{"github.com/sharasha07/royale-tourneys"},
	}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return string(jwtBytes), nil
}

func NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func tokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func (m DBModel) AddRefreshToken(ctx context.Context, token string, userID int, ttl time.Duration) error {
	query := `
		INSERT INTO refresh_tokens(token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	expiry := time.Now().Add(ttl)

	_, err := m.pool.Exec(ctx, query, tokenHash(token), userID, expiry)
	return err
}

func (m DBModel) GetRefreshTokenUserID(ctx context.Context, token string) (int, error) {
	query := `
		SELECT user_id FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var userID int
	err := m.pool.QueryRow(ctx, query, tokenHash(token)).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (m DBModel) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.pool.Exec(ctx, query, tokenHash(token))
	return err
}

func (m DBModel) DeleteAllRefreshToken(ctx context.Context, userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.pool.Exec(ctx, query, userID)
	return err
}
