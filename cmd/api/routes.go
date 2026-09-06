package main

import (
	"expvar"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("GET /debug/vars", expvar.Handler())

	mux.HandleFunc("POST /v1/users", app.createUserHandler)
	mux.HandleFunc("POST /v1/users/login", app.loginHandler)
	mux.HandleFunc("POST /v1/users/token/refresh", app.refreshTokenHandler)
	mux.HandleFunc("GET /v1/users/{id}", app.showUserHandler)
	mux.HandleFunc("PATCH /v1/users/{id}", app.updateUserHandler)
	mux.HandleFunc("PUT /v1/users/{id}/tag", app.updateGameTagHandler)
	mux.HandleFunc("PUT /v1/users/{id}/profile_picture", app.updateProfilePictureHandler)
	mux.HandleFunc("DELETE /v1/users/{id}", app.deleteUserHandler)

	return app.metrics(recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(mux)))))
}
