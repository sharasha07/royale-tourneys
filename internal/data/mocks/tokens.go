package mocks

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/sharasha07/royale-tourneys/internal/data"
)

type TokenModel struct {
	tokens map[string]data.Token
}

func NewTokenModel() *TokenModel {
	return &TokenModel{
		tokens: map[string]data.Token{
			"token_example": {
				TokenHash: tokenHash("token_example"),
				UserID:    1,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				CreatedAt: time.Now(),
			},
		},
	}
}

func tokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func (m *TokenModel) Insert(ctx context.Context, token string, userID int, ttl time.Duration) error {
	m.tokens[token] = data.Token{
		TokenHash: tokenHash(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	return nil
}

func (m *TokenModel) GetUserID(ctx context.Context, token string) (int, error) {
	t, ok := m.tokens[token]
	if !ok {
		return 0, data.ErrNoRecord
	}

	if time.Now().After(t.ExpiresAt) {
		return 0, data.ErrNoRecord
	}

	return t.UserID, nil
}

func (m *TokenModel) Delete(ctx context.Context, token string) error {
	delete(m.tokens, token)

	return nil
}
