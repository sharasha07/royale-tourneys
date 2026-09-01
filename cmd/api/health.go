package main

import "net/http"

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"})
	if err != nil {
		app.serverErrorResponse(w, err)
	}
}
