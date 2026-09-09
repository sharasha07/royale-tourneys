package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sharasha07/royale-tourneys/internal/data"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, &input.Username, &input.Password)

	if ok := v.Valid(); !ok {
		failedValidationResponse(w, v.Errors)
		return
	}

	user, err := app.models.Users.Insert(r.Context(), input.Username, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUniqueViolation):
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
		notFoundResponse(w)
		return
	}

	user, err := app.models.Users.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
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
		notFoundResponse(w)
		return
	}

	if id != user.ID {
		notAuthorizedResponse(w)
		return
	}

	var input struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}

	err = readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
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

	if input.Username == nil && input.Password == nil {
		v := validator.New()
		v.Add("body", "must not be empty")
		failedValidationResponse(w, v.Errors)
		return
	}

	err = app.models.Users.Update(r.Context(), user)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, data.ErrEditConflict):
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
		notFoundResponse(w)
		return
	}

	if id != user.ID {
		notAuthorizedResponse(w)
		return
	}

	var input struct {
		GameTag string `json:"game_tag"`
	}

	err = readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	v := validator.New()
	err = data.ValidateGameTag(v, input.GameTag, app.cfg.ClashAPIToken, app.httpClient)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	if ok := v.Valid(); !ok {
		failedValidationResponse(w, v.Errors)
		return
	}

	user.GameTag = &input.GameTag

	err = app.models.Users.Update(r.Context(), user)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, data.ErrEditConflict):
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
		notFoundResponse(w)
		return
	}

	if id != user.ID {
		notAuthorizedResponse(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		if mbErr, ok := errors.AsType[*http.MaxBytesError](err); ok {
			log.Println("upload too large, limit:", mbErr.Limit)
			fileTooLargeResponse(w)
			return
		}
		badRequestResponse(w, err)
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		serverErrorResponse(w, err)
		return
	}
	contentType := http.DetectContentType(buf[:n])

	var key string
	switch contentType {
	case "image/jpeg":
		key = fmt.Sprintf("users/%d/profile_picture.jpg", id)
	case "image/png":
		key = fmt.Sprintf("users/%d/profile_picture.png", id)
	case "image/webp":
		key = fmt.Sprintf("users/%d/profile_picture.webp", id)
	default:
		badRequestResponse(w, errors.New("invalid Content-Type"))
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		serverErrorResponse(w, err)
		return
	}

	_, err = app.s3Client.PutObject(r.Context(),
		&s3.PutObjectInput{
			Bucket:       &app.cfg.R2.Bucket,
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

	endpoint := path.Join(app.cfg.R2.PublicURL, key)
	user.ProfilePicture = &endpoint

	err = app.models.Users.Update(r.Context(), user)
	if err != nil {
		_, delErr := app.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(app.cfg.R2.Bucket),
			Key:    aws.String(key),
		})

		if delErr != nil {
			log.Println("failed to remove orphaned profile picture:", delErr)
		}

		switch {
		case errors.Is(err, data.ErrEditConflict):
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
		notFoundResponse(w)
		return
	}

	if id != user.ID {
		notAuthorizedResponse(w)
		return
	}

	err = app.models.Users.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNoRecord):
			notFoundResponse(w)
		default:
			serverErrorResponse(w, err)
		}
		return
	}

	if user.ProfilePicture != nil {
		ext := path.Ext(*user.ProfilePicture)
		key := fmt.Sprintf("users/%d/profile_picture%s", id, ext)

		_, delErr := app.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(app.cfg.R2.Bucket),
			Key:    aws.String(key),
		})

		if delErr != nil {
			log.Println("failed to remove profile picture:", delErr)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
