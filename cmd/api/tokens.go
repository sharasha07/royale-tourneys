package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
	"github.com/pascaldekloe/jwt"
	"github.com/sharasha07/royale-tourneys/internal/data"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
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
	data.ValidateUser(v, &input.Username, &input.Password)

	if ok := v.Valid(); !ok {
		failedValidationResponse(w, v.Errors)
		return
	}

	user, err := app.model.GetUserByUsername(input.Username)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			invalidCredentialsResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	match, err := argon2id.ComparePasswordAndHash(input.Password, string(user.PasswordHash))
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	if !match {
		invalidCredentialsResponse(w)
		return
	}

	claims := jwt.Claims{
		Subject:   strconv.FormatInt(int64(user.ID), 10),
		Issued:    jwt.NewNumericTime(time.Now()),
		NotBefore: jwt.NewNumericTime(time.Now()),
		Expires:   jwt.NewNumericTime(time.Now().Add(24 * time.Hour)),
		Issuer:    "github.com/sharasha07/royale-tourneys",
		Audiences: []string{"github.com/sharasha07/royale-tourneys"},
	}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.cfg.jwtSecret))
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, envelope{"authentication_token": string(jwtBytes)})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}
