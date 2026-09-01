package main

import "net/http"

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	message := "status: available"

	panic("blabla")
	w.Write([]byte(message))
}
