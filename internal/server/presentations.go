package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/PabloVarg/presentation-timer/internal/helpers"
	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/PabloVarg/presentation-timer/internal/validation"
)

func CreatePresentationHandler(logger *slog.Logger, queries *queries.Queries) http.Handler {
	type Input struct {
		Name string `json:"name"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input Input

		if err := helpers.ReadJSON(r.Body, &input); err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		v := validation.New()
		v.Check(
			"name",
			input.Name,
			validation.CheckNotEmpty("name can't be empty"),
			validation.CheckLength(5, 50, "name must be between 5 and 50 characters"),
		)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		presentation, err := queries.CreatePresentation(ctx, input.Name)
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		if err := helpers.WriteJSON(w, http.StatusCreated, presentation); err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}
