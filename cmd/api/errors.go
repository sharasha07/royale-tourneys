package main

import (
	"log"
	"net/http"
)

func (app *application) sendError(w http.ResponseWriter, status int, message any) {
	err := app.writeJSON(w, status, envelope{"error": message})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, err error) {
	log.Println(err)

	message := "internal server error"
	app.sendError(w, http.StatusInternalServerError, message)
}
