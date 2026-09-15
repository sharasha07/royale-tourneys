package data

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
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

func (m TournamentModel) GetAll(ctx context.Context, id int, name string, filters Filters) ([]Tournament, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, max_players, user_id, created_at, version
		FROM tournaments
		WHERE (id = $1 OR $1 = 0)
		AND (LOWER(name) = LOWER($2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []any{id, name, filters.limit(), filters.offset()}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := m.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var tournaments []Tournament

	for rows.Next() {
		var tournament Tournament

		err := rows.Scan(
			&tournament.ID,
			&tournament.Name,
			&tournament.Description,
			&tournament.MaxPlayers,
			&tournament.UserID,
			&tournament.CreatedAt,
			&tournament.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		tournaments = append(tournaments, tournament)
	}

	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(len(tournaments), filters.Page, filters.PageSize)

	return tournaments, metadata, nil
}

func (m TournamentModel) GetByID(ctx context.Context, id int) (Tournament, error) {
	query := `
		SELECT id, name, description, max_players, user_id, created_at, version
		FROM tournaments
		WHERE id = $1`

	var tournament Tournament

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := m.pool.QueryRow(ctx, query, id).Scan(
		&tournament.ID,
		&tournament.Name,
		&tournament.Description,
		&tournament.MaxPlayers,
		&tournament.UserID,
		&tournament.CreatedAt,
		&tournament.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return Tournament{}, ErrNoRecord
		default:
			return Tournament{}, err
		}
	}

	return tournament, nil
}

func (m TournamentModel) GetByUserID(ctx context.Context, userID int) ([]Tournament, error) {
	query := `
		SELECT id, name, description, max_players, created_at, version
		FROM tournaments
		WHERE user_id = $1`

	var tournaments []Tournament

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := m.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		tournament := Tournament{UserID: userID}

		err := rows.Scan(
			&tournament.ID,
			&tournament.Name,
			&tournament.Description,
			&tournament.MaxPlayers,
			&tournament.CreatedAt,
			&tournament.Version,
		)
		if err != nil {
			return nil, err
		}

		tournaments = append(tournaments, tournament)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tournaments, nil
}

func (m TournamentModel) Update(ctx context.Context, tournament *Tournament) error {
	query := `
		UPDATE tournaments
		SET name = $1, description = $2, password_hash = $3, max_players = $4, version = version + 1
		WHERE id = $5 and version = $6
		RETURNING version`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := []any{tournament.Name, tournament.Description, tournament.PasswordHash, tournament.MaxPlayers, tournament.ID, tournament.Version}
	err := m.pool.QueryRow(ctx, query, args...).Scan(
		&tournament.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m TournamentModel) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM tournaments
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.pool.Exec(ctx, query, id)

	return err
}
