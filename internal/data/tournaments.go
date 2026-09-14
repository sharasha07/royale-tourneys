package data

import (
	"context"
	"fmt"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

type TournamentModel struct {
	pool *pgxpool.Pool
}

type Tournament struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	PasswordHash *string   `json:"-"`
	MaxPlayers   int       `json:"max_players"`
	UserID       int       `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	Version      int       `json:"version"`
}

func ValidateTournament(v *validator.Validator, name, description, password *string, maxPlayers *int) {
	if name != nil {
		v.Check(*name != "", "name", "must be provided")
		v.Check(utf8.RuneCountInString(*name) <= 10, "name", "must not be more than 10 characters")
	}

	if description != nil {
		v.Check(utf8.RuneCountInString(*description) <= 100, "description", "must not be more than 100 characters")
	}

	if password != nil {
		v.Check(utf8.RuneCountInString(*password) > 5, "password", "must be more than 5 characters")
		v.Check(utf8.RuneCountInString(*password) <= 15, "password", "must not be more than 15 characters")
	}

	if maxPlayers != nil {
		validMaxPlayers := []int{10, 20, 30, 40}
		v.Check(slices.Contains(validMaxPlayers, *maxPlayers), "max_players", fmt.Sprintf("must be in: %v", validMaxPlayers))
	}
}

func (m TournamentModel) Insert(ctx context.Context, name string, description, password *string, maxPlayers int, userID int) (Tournament, error) {
	query := `
		INSERT INTO tournaments(name, description, password_hash, max_players, user_id)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, created_at, version`

	t := Tournament{
		Name:        name,
		Description: description,
		MaxPlayers:  maxPlayers,
		UserID:      userID,
	}

	if password != nil {
		hash, err := argon2id.CreateHash(*password, argon2id.DefaultParams)
		if err != nil {
			return Tournament{}, err
		}

		t.PasswordHash = &hash
	}

	args := []any{t.Name, t.Description, t.PasswordHash, t.MaxPlayers, t.UserID}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := m.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID,
		&t.CreatedAt,
		&t.Version,
	)
	if err != nil {
		return Tournament{}, err
	}

	return t, nil
}
