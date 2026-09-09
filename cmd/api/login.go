package main

import (
	"errors"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/sharasha07/royale-tourneys/internal/data"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	user, err := app.models.Users.GetByUsername(r.Context(), input.Username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
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

	jwtToken, err := data.NewAccessToken(user.ID, app.cfg.JWT.Secret, app.cfg.JWT.AccessTTL)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	refreshToken, err := data.NewRefreshToken()
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = app.models.Tokens.Insert(r.Context(), refreshToken, user.ID, app.cfg.JWT.RefreshTTL)
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

	err := readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	if input.RefreshToken == "" {
		v := validator.New()
		v.Add("refresh_token", "must be provided")
		failedValidationResponse(w, v.Errors)
		return
	}

	userID, err := app.models.Tokens.GetUserID(r.Context(), input.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			invalidAuthenticationTokenResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	newAccessToken, err := data.NewAccessToken(userID, app.cfg.JWT.Secret, app.cfg.JWT.AccessTTL)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{"access_token": newAccessToken})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	err := readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	if input.RefreshToken == "" {
		v := validator.New()
		v.Add("refresh_token", "must be provided")
		failedValidationResponse(w, v.Errors)
		return
	}

	err = app.models.Tokens.Delete(r.Context(), input.RefreshToken)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
