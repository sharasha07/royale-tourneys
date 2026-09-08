package mocks

import (
	"context"
	"time"

	"github.com/sharasha07/royale-tourneys/internal/data"
)

var mockUser = data.User{
	ID:        1,
	Username:  "shaba",
	CreatedAt: time.Date(2003, time.July, 1, 22, 0, 0, 0, time.UTC),
	Version:   2,
}

type UserModel struct{}

func (m UserModel) Insert(ctx context.Context, username, password string) (data.User, error) {
	return data.User{}, nil
}

func (m UserModel) GetByID(ctx context.Context, id int) (data.User, error) {
	switch id {
	case 1:
		return mockUser, nil
	default:
		return data.User{}, data.ErrNoRecord
	}
}

func (m UserModel) GetByUsername(ctx context.Context, username string) (data.User, error) {
	return data.User{}, nil
}

func (m UserModel) Update(ctx context.Context, user *data.User) error {
	return nil
}

func (m UserModel) Delete(ctx context.Context, id int) error {
	return nil
}
