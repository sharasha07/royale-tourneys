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

func badRequestResponse(w http.ResponseWriter) {
	message := "bad request"
	sendError(w, http.StatusBadRequest, message)
}

func notFoundResponse(w http.ResponseWriter) {
	message := "resource not found"
	sendError(w, http.StatusNotFound, message)
}

func failedValidationResponse(w http.ResponseWriter, errors map[string]string) {
	sendError(w, http.StatusUnprocessableEntity, errors)
}

func editConflictResponse(w http.ResponseWriter) {
	message := "edit conflict"
	sendError(w, http.StatusConflict, message)
}

func invalidCredentialsResponse(w http.ResponseWriter) {
	message := "invalid credentials"
	sendError(w, http.StatusUnauthorized, message)
}

func invalidAuthenticationTokenResponse(w http.ResponseWriter) {
	message := "invalid authentication token"
	sendError(w, http.StatusUnauthorized, message)
}
