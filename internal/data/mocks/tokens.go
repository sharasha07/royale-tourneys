package mocks

import (
	"context"
	"time"
)

type TokenModel struct{}

func (m TokenModel) Insert(ctx context.Context, token string, userID int, ttl time.Duration) error {
	return nil
}

func (m TokenModel) GetUserID(ctx context.Context, token string) (int, error) {
	return 0, nil
}

func (m TokenModel) Delete(ctx context.Context, token string) error {
	return nil
}

func (m TokenModel) DeleteAllForUser(ctx context.Context, userID int) error {
	return nil
}
