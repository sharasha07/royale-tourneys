package main

import (
	"net/http"

	"github.com/sharasha07/royale-tourneys/internal/data"
	"github.com/sharasha07/royale-tourneys/internal/validator"
)

func (app *application) createTournamentHandler(w http.ResponseWriter, r *http.Request) {
	user := contextGetUser(r)

	var input struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Password    *string `json:"password"`
		MaxPlayers  int     `json:"max_players"`
	}

	err := readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	v := validator.New()
	if data.ValidateTournament(v, &input.Name, input.Description, input.Password, &input.MaxPlayers); !v.Valid() {
		failedValidationResponse(w, v.Errors)
		return
	}

	tournament, err := app.models.Tournaments.Insert(
		r.Context(), input.Name, input.Description, input.Password, input.MaxPlayers, user.ID)

	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, envelope{"tournament": tournament})
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}
