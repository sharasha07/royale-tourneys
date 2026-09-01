package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func (app *application) readJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	return dec.Decode(dst)
}

func (app *application) writeJSON(w http.ResponseWriter, status int, env envelope) error {
	data, err := json.MarshalIndent(env, "", "\t")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)

	return nil
}
