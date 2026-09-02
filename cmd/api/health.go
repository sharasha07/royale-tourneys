package main

import "net/http"

func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := writeJSON(w, http.StatusOK, envelope{"status": "available"})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}
