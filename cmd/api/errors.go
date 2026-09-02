package main

import (
	"log"
	"net/http"
)

func sendError(w http.ResponseWriter, status int, message any) {
	err := writeJSON(w, status, envelope{"error": message})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func serverErrorResponse(w http.ResponseWriter, err error) {
	log.Println(err)

	message := "internal server error"
	sendError(w, http.StatusInternalServerError, message)
}
