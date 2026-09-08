package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoRecord     = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Users  UserModelInterface
	Tokens TokenModelInterface
}

type UserModelInterface interface {
	Insert(ctx context.Context, username, password string) (User, error)
	GetByID(ctx context.Context, id int) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
}

type TokenModelInterface interface {
	Insert(ctx context.Context, token string, userID int, ttl time.Duration) error
	GetUserID(ctx context.Context, token string) (int, error)
	Delete(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID int) error
}

func NewDBModels(pool *pgxpool.Pool) Models {
	return Models{
		Users:  UserModel{pool: pool},
		Tokens: TokenModel{pool: pool},
	}
}
