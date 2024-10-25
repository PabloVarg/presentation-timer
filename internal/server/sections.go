package server

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/PabloVarg/presentation-timer/internal/helpers"
	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/PabloVarg/presentation-timer/internal/validation"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreateSectionHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type input struct {
		Presentation *int64         `json:"presentation"`
		Name         *string        `json:"name"`
		Duration     *time.Duration `json:"duration"`
		Position     *int16         `json:"position"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input input

		if err := helpers.ReadJSON(r.Body, &input); err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		v := validation.New()
		v.Check(
			"presentation",
			input.Presentation,
			validation.CheckPointerNotNil("presentation must be given"),
			validation.IntCheckNatural("presentation can not be negative"),
		)
		v.Check(
			"name",
			input.Name,
			validation.CheckPointerNotNil("name must be given"),
			validation.StringCheckNotEmpty("name can not be empty"),
			validation.StringCheckLength(5, 50, "name must be between 5 and 50 characters"),
		)
		v.Check(
			"duration",
			input.Duration,
			validation.CheckPointerNotNil("duration must be given"),
			validation.DurationCheckPositive("duration can not be negative"),
			validation.DurationCheckMin("duration can not be less than 1 second", time.Second),
		)
		v.Check(
			"position",
			input.Position,
			validation.CheckPointerNotNil("position must be given"),
			validation.IntCheckNatural("position can not be negative"),
		)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		presentation, err := queriesStore.CreateSection(ctx, queries.CreateSectionParams{
			Presentation: *input.Presentation,
			Name:         *input.Name,
			Duration:     *input.Duration,
			Position:     *input.Position,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			switch errors.As(err, &pgErr) {
			case pgErr.ConstraintName != "":
				v := validation.New()
				v.AddErrors("presentation", "presentation does not exist")
				helpers.UnprocessableContent(w, v.Errors())
			default:
				helpers.InternalError(w, logger, err)
			}
			return
		}

		if err := helpers.WriteJSON(w, http.StatusCreated, presentation); err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}

func DeleteSectionHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := queriesStore.DeleteSection(ctx, ID)
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
		if rows == 0 {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
