package main

import (
	"errors"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := readJSON(r, &input)
	if err != nil {
		badRequestResponse(w)
		return
	}

	v := validator.New()
	v.Check(input.Username != "", "username", "must be provided")
	v.Check(utf8.RuneCountInString(input.Username) <= 10, "username", "must not have more than 10 characters")

	v.Check(input.Password != "", "password", "must be provided")
	v.Check(utf8.RuneCountInString(input.Password) > 5, "password", "must be more than 5 characters")

	if ok := v.Valid(); !ok {
		failedValidationResponse(w, v.Errors)
		return
	}

	hash, err := argon2id.CreateHash(input.Password, argon2id.DefaultParams)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	user, err := app.model.CreateUser(input.Username, []byte(hash))
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			v.Add("username", "must be unique")
			failedValidationResponse(w, v.Errors)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, envelope{"user": user})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) showUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		badRequestResponse(w)
		return
	}

	user, err := app.model.GetUserByID(id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			notFoundResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{"user": user})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}
