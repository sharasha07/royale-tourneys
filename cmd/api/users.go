package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sharasha07/royale-tourneys/internal/data"
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
	data.ValidateUser(v, &input.Username, &input.Password)

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

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	user := contextGetUser(r)

	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		badRequestResponse(w)
		return
	}

	if id != user.ID {
		unauthorizedResponse(w)
		return
	}

	var input struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}

	err = readJSON(r, &input)
	if err != nil {
		badRequestResponse(w)
		return
	}

	v := validator.New()
	data.ValidateUser(v, input.Username, input.Password)
	if ok := v.Valid(); !ok {
		failedValidationResponse(w, v.Errors)
		return
	}

	if input.Username != nil {
		user.Username = *input.Username
	}

	if input.Password != nil {
		hash, err := argon2id.CreateHash(*input.Password, argon2id.DefaultParams)
		if err != nil {
			serverErrorResponse(w, err)
			return
		}

		user.PasswordHash = []byte(hash)
	}

	err = app.model.UpdateUser(user)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			editConflictResponse(w)
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			v.Add("username", "must be unique")
			failedValidationResponse(w, v.Errors)
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

func (app *application) updateGameTagHandler(w http.ResponseWriter, r *http.Request) {
	user := contextGetUser(r)

	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		badRequestResponse(w)
		return
	}

	if id != user.ID {
		unauthorizedResponse(w)
		return
	}

	var input struct {
		GameTag string `json:"game_tag"`
	}

	err = readJSON(r, &input)
	if err != nil {
		badRequestResponse(w)
		return
	}

	v := validator.New()
	err = data.ValidateGameTag(v, input.GameTag, app.cfg.clashAPIToken, app.client)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	if ok := v.Valid(); !ok {
		failedValidationResponse(w, v.Errors)
		return
	}

	user.GameTag = &input.GameTag

	err = app.model.UpdateUser(user)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			editConflictResponse(w)
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			v.Add("game_tag", "must be unique")
			failedValidationResponse(w, v.Errors)
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

func (app *application) updateProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	user := contextGetUser(r)

	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		badRequestResponse(w)
		return
	}

	if id != user.ID {
		unauthorizedResponse(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		badRequestResponse(w)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		badRequestResponse(w)
		return
	}
	defer file.Close()

	var key string
	contentType := header.Header.Get("Content-Type")
	switch contentType {
	case "":
		badRequestResponse(w)
		return
	case "image/jpeg":
		key = fmt.Sprintf("users/%d/profile_picture.jpg", id)
	case "image/png":
		key = fmt.Sprintf("users/%d/profile_picture.png", id)
	case "image/webp":
		key = fmt.Sprintf("users/%d/profile_picture.webp", id)
	default:
		badRequestResponse(w)
		return
	}

	_, err = app.s3Client.PutObject(r.Context(),
		&s3.PutObjectInput{
			Bucket:       &app.cfg.r2Bucket,
			Key:          &key,
			Body:         file,
			ContentType:  &contentType,
			CacheControl: aws.String("public, max-age=3600"),
		},
	)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	endpoint := app.cfg.r2PublicURL + "/" + key
	user.ProfilePicture = &endpoint

	err = app.model.UpdateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			editConflictResponse(w)
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

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := contextGetUser(r)

	if user.IsAnonymous() {
		authenticationRequiredResponse(w)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		badRequestResponse(w)
		return
	}

	if id != user.ID {
		unauthorizedResponse(w)
		return
	}

	err = app.model.DeleteUser(id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			notFoundResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
