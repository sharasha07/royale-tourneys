package main

import (
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func (app *application) showUsersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		username string
		tag      string
		data.Filters
	}

	qs := r.URL.Query()
	v := validator.New()

	input.username = readString(qs, "username", "")
	input.tag = readString(qs, "tag", "")

	page, err := readInt(qs, "page", 1)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	pageSize, err := readInt(qs, "page_size", 20)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	input.Filters.Page = page
	input.Filters.PageSize = pageSize
	input.Filters.Sort = readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "-id", "username", "-username"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		failedValidationResponse(w, v.Errors)
		return
	}

	users, metadata, err := app.models.Users.GetAll(r.Context(), input.username, input.tag, input.Filters)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "users": users})
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
		forbiddenResponse(w)
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

	if input.Username == nil && input.Password == nil {
		v := validator.New()
		v.Add("fields", "at least one should be non-null")
		failedValidationResponse(w, v.Errors)
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

		user.PasswordHash = hash
	}

	err = app.models.Users.Update(r.Context(), user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			editConflictResponse(w)
		case errors.Is(err, data.ErrUniqueViolation):
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
		forbiddenResponse(w)
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
	err = data.ValidateGameTag(r.Context(), v, input.GameTag, app.cfg.ClashAPIToken, app.httpClient)
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
		switch {
		case errors.Is(err, data.ErrEditConflict):
			editConflictResponse(w)
		case errors.Is(err, data.ErrUniqueViolation):
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
		forbiddenResponse(w)
		return
	}

	const maxRequestSize = 5 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	err = r.ParseMultipartForm(maxRequestSize)
	if err != nil {
		badRequestResponse(w, errors.New("malformed upload"))
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrMissingFile):
			badRequestResponse(w, errors.New("avatar file is required"))
		default:
			badRequestResponse(w, errors.New("malformed upload"))
		}
		return
	}
	defer file.Close()

	_, format, err := image.DecodeConfig(file)
	if err != nil {
		badRequestResponse(w, errors.New("invalid image"))
		return
	}

	supportedFormats := []string{"jpeg", "png", "webp"}
	if !slices.Contains(supportedFormats, format) {
		badRequestResponse(w, errors.New("unsupported image type"))
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		serverErrorResponse(w, err)
		return
	}

	key := fmt.Sprintf("users/%d/profile_picture", id)

	endpoint, err := url.JoinPath(app.cfg.R2.PublicURL, key)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	endpoint += "?v=" + strconv.FormatInt(time.Now().UnixNano(), 10)

	_, err = app.s3Client.PutObject(r.Context(),
		&s3.PutObjectInput{
			Bucket:       aws.String(app.cfg.R2.Bucket),
			Key:          aws.String(key),
			Body:         file,
			ContentType:  aws.String("image/" + format),
			CacheControl: aws.String("public, max-age=3600"),
		},
	)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	user.ProfilePicture = &endpoint

	err = app.models.Users.Update(r.Context(), user)
	if err != nil {
		_, delErr := app.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(app.cfg.R2.Bucket),
			Key:    aws.String(key),
		})

		if delErr != nil {
			log.Println("failed to remove orphaned profile picture: ", delErr)
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
		forbiddenResponse(w)
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
