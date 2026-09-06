package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pascaldekloe/jwt"
	"github.com/sharasha07/royale-tourneys/internal/data"
)

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				log.Println(err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			r = contextSetUser(r, data.AnonymousUser)

			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			invalidAuthenticationTokenResponse(w)
			return
		}

		token := parts[1]

		claims, err := jwt.HMACCheck([]byte(token), []byte(app.cfg.jwt.secret))
		if err != nil {
			invalidAuthenticationTokenResponse(w)
			return
		}

		if !claims.Valid(time.Now()) || claims.Issuer != "github.com/sharasha07/royale-tourneys" || !claims.AcceptAudience("github.com/sharasha07/royale-tourneys") {
			invalidAuthenticationTokenResponse(w)
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			serverErrorResponse(w, err)
			return
		}

		user, err := app.model.GetUserByID(int(userID))
		if err != nil {
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				invalidAuthenticationTokenResponse(w)
			default:
				serverErrorResponse(w, err)
			}
			return
		}

		r = contextSetUser(r, &user)

		next.ServeHTTP(w, r)
	})
}
