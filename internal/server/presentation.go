package server

import (
	"log/slog"
	"net/http"

	"github.com/PabloVarg/presentation-timer/internal/helpers"
)

func CreatePresentationHandler(logger *slog.Logger) http.Handler {
	type Input struct {
		Name string `json:"name"`
	}

	type Output struct {
		Received Input `json:"received"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input Input

		if err := helpers.ReadJSON(r.Body, &input); err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		helpers.WriteJSON(w, http.StatusOK, Output{
			Received: input,
		})
	})
}
