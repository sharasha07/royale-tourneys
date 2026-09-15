package main

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/sharasha07/royale-tourneys/internal/data"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

func (app *application) createTournamentHandler(w http.ResponseWriter, r *http.Request) {
	user := contextGetUser(r)
	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	var input struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Password    *string `json:"password"`
		MaxPlayers  int     `json:"max_players"`
	}

	err := readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	v := validator.New()
	if data.ValidateTournament(v, &input.Name, input.Description, input.Password, &input.MaxPlayers); !v.Valid() {
		failedValidationResponse(w, v.Errors)
		return
	}

	tournament, err := app.models.Tournaments.Insert(
		r.Context(), input.Name, input.Description, input.Password, input.MaxPlayers, user.ID)

	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, envelope{"tournament": tournament})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) showTournamentsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		id   int
		name string
		data.Filters
	}

	qs := r.URL.Query()
	v := validator.New()

	n, err := readInt(qs, "id", 0)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	input.id = n
	input.name = readString(qs, "name", "")

	page, err := readInt(qs, "page", 1)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	page_size, err := readInt(qs, "page_size", 20)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	input.Filters.Page = page
	input.Filters.PageSize = page_size
	input.Filters.Sort = readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "-id", "name", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		failedValidationResponse(w, v.Errors)
		return
	}

	tournaments, metadata, err := app.models.Tournaments.GetAll(r.Context(), input.id, input.name, input.Filters)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{
		"metadata":    metadata,
		"tournaments": tournaments,
	})

	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) showTournamentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		notFoundResponse(w)
		return
	}

	tournament, err := app.models.Tournaments.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			notFoundResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{"tournament": tournament})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) updateTournamentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		notFoundResponse(w)
		return
	}

	user := contextGetUser(r)
	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	tournaments, err := app.models.Tournaments.GetByUserID(r.Context(), user.ID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	if !slices.ContainsFunc(tournaments, func(t data.Tournament) bool {
		return t.ID == id
	}) {
		forbiddenResponse(w)
		return
	}

	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Password    *string `json:"password"`
		MaxPlayers  *int    `json:"max_players"`
	}

	err = readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	v := validator.New()
	if data.ValidateTournament(v, input.Name, input.Description, input.Password, input.MaxPlayers); !v.Valid() {
		failedValidationResponse(w, v.Errors)
		return
	}

	i := slices.IndexFunc(tournaments, func(t data.Tournament) bool {
		return t.ID == id
	})
	tournament := tournaments[i]

	if input.Name != nil {
		tournament.Name = *input.Name
	}

	if input.Description != nil {
		tournament.Description = input.Description
	}

	if input.Password != nil {
		hash, err := argon2id.CreateHash(*input.Password, argon2id.DefaultParams)
		if err != nil {
			serverErrorResponse(w, err)
			return
		}

		tournament.PasswordHash = &hash
	}

	if input.MaxPlayers != nil {
		tournament.MaxPlayers = *input.MaxPlayers
	}

	err = app.models.Tournaments.Update(r.Context(), &tournament)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			editConflictResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{"tournament": tournament})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) deleteTournamentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		notFoundResponse(w)
		return
	}

	user := contextGetUser(r)
	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	tournaments, err := app.models.Tournaments.GetByUserID(r.Context(), user.ID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	if !slices.ContainsFunc(tournaments, func(t data.Tournament) bool {
		return t.ID == id
	}) {
		forbiddenResponse(w)
		return
	}

	err = app.models.Tournaments.Delete(r.Context(), id)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
