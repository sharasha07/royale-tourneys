package main

import (
	"errors"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := readJSON(r, &input)
	if err != nil {
		badRequestResponse(w)
		return
	}

	user, err := app.model.GetUserByUsername(input.Username)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			notFoundResponse(w)
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

	jwtToken, err := data.NewJWTToken(user.ID, app.cfg.jwtSecret, app.cfg.jwtAccessTTL)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	refreshToken, err := data.NewRefreshToken()
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = app.model.AddRefreshToken(refreshToken, user.ID, app.cfg.jwtRefreshTTL)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{
		"access_token":  jwtToken,
		"refresh_token": refreshToken,
	})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := readJSON(r, &input)
	if err != nil || input.RefreshToken == "" {
		badRequestResponse(w)
		return
	}

	userID, err := app.model.GetRefreshTokenUserID(input.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			invalidAuthenticationTokenResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	err = app.model.DeleteRefreshToken(input.RefreshToken)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	newAccessToken, err := data.NewJWTToken(userID, app.cfg.jwtSecret, app.cfg.jwtAccessTTL)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	newRefreshToken, err := data.NewRefreshToken()
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = app.model.AddRefreshToken(newRefreshToken, userID, app.cfg.jwtRefreshTTL)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}
