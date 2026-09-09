package mocks

import (
	"context"
	"slices"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

var mockUsers = []data.User{
	data.User{
		ID:        1,
		Username:  "shaba",
		CreatedAt: time.Now(),
		Version:   2,
	},
}

type UserModel struct{}

func (m UserModel) Insert(ctx context.Context, username, password string) (data.User, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return data.User{}, err
	}

	u := data.User{
		ID:             len(mockUsers),
		Username:       username,
		PasswordHash:   []byte(hash),
		GameTag:        nil,
		ProfilePicture: nil,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	if slices.ContainsFunc(mockUsers, func(user data.User) bool { return user.Username == username }) {
		return data.User{}, data.ErrUniqueViolation
	}

	mockUsers = append(mockUsers, u)

	return u, nil
}

func (m UserModel) GetByID(ctx context.Context, id int) (data.User, error) {
	return data.User{}, data.ErrNoRecord
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
