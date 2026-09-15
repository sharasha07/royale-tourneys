package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoRecord            = errors.New("record not found")
	ErrEditConflict        = errors.New("edit conflict")
	ErrUniqueViolation     = errors.New("unique violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrTournamentFull      = errors.New("tournament is full")
)

type Models struct {
	Users       UserModelInterface
	Tokens      TokenModelInterface
	Tournaments TournamentModelInterface
}

type UserModelInterface interface {
	Insert(ctx context.Context, username, password string) (User, error)
	GetAll(ctx context.Context, username, tag string, filters Filters) ([]User, Metadata, error)
	GetByID(ctx context.Context, id int) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
}

type TokenModelInterface interface {
	Insert(ctx context.Context, token string, userID int, ttl time.Duration) error
	GetUserID(ctx context.Context, token string) (int, error)
	Delete(ctx context.Context, token string) error
}

type TournamentModelInterface interface {
	Insert(ctx context.Context, name string, description, password *string, maxPlayers int, userID int) (Tournament, error)
	GetAll(ctx context.Context, id int, name string, filters Filters) ([]Tournament, Metadata, error)
	GetByID(ctx context.Context, id int) (Tournament, error)
	GetByUserID(ctx context.Context, userID int) ([]Tournament, error)
	Update(ctx context.Context, tournament *Tournament) error
	Delete(ctx context.Context, id int) error
	AddUser(ctx context.Context, tour_id, user_id int) error
	RemoveUser(ctx context.Context, tour_id, user_id int) error
}

func NewDBModels(pool *pgxpool.Pool) Models {
	return Models{
		Users:       UserModel{pool: pool},
		Tokens:      TokenModel{pool: pool},
		Tournaments: TournamentModel{pool: pool},
	}
}
