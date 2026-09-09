package mocks

import (
	"context"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

type UserModel struct {
	users  map[int]data.User
	nextID int
}

func NewUserModel() *UserModel {
	hash, err := argon2id.CreateHash("shaba123", argon2id.DefaultParams)
	if err != nil {
		panic(err)
	}

	return &UserModel{
		users: map[int]data.User{
			1: {
				ID:           1,
				Username:     "shaba",
				PasswordHash: []byte(hash),
				CreatedAt:    time.Now(),
				Version:      2,
			},
		},
		nextID: 2,
	}
}

func (m *UserModel) Insert(ctx context.Context, username, password string) (data.User, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return data.User{}, err
	}

	for _, user := range m.users {
		if user.Username == username {
			return data.User{}, data.ErrUniqueViolation
		}
	}

	u := data.User{
		ID:             m.nextID,
		Username:       username,
		PasswordHash:   []byte(hash),
		GameTag:        nil,
		ProfilePicture: nil,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	m.users[u.ID] = u
	m.nextID++

	return u, nil
}

func (m *UserModel) GetByID(ctx context.Context, id int) (data.User, error) {
	user, ok := m.users[id]
	if !ok {
		return data.User{}, data.ErrNoRecord
	}

	return user, nil
}

func (m *UserModel) GetByUsername(ctx context.Context, username string) (data.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}

	return data.User{}, data.ErrNoRecord
}

func (m *UserModel) Update(ctx context.Context, user *data.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return data.ErrNoRecord
	}

	user.Version++
	m.users[user.ID] = *user

	return nil
}

func (m *UserModel) Delete(ctx context.Context, id int) error {
	if _, ok := m.users[id]; !ok {
		return data.ErrNoRecord
	}

	delete(m.users, id)

	return nil
}
