package server

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/PabloVarg/presentation-timer/internal/filters"
	"github.com/PabloVarg/presentation-timer/internal/helpers"
	"github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/PabloVarg/presentation-timer/internal/validation"
)

const defaultPageSize = 20

var validSortByFields = []string{"name"}

func ListPresentationsHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, v := filters.FromRequest(r, defaultPageSize, validSortByFields...)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		presentations, err := queriesStore.GetPresentations(ctx, queries.GetPresentationsParams{
			Direction:   f.QuerySortDirection(),
			SortBy:      f.QuerySortBy(),
			QueryOffset: f.QueryOffset(),
			QueryLimit:  f.QueryLimit(),
		})
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		if err := helpers.WriteJSON(w, http.StatusOK, presentations); err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}

func GetPresentationHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		presentation, err := queriesStore.GetPresentation(ctx, ID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.NotFound(w, r)
			default:
				helpers.InternalError(w, logger, err)
			}
			return
		}

		if err := helpers.WriteJSON(w, http.StatusOK, presentation); err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}

func CreatePresentationHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type Input struct {
		Name *string `json:"name"`
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
			validation.CheckPointerNotNil("name must be given"),
			validation.StringCheckNotEmpty("name can't be empty"),
			validation.StringCheckLength(5, 50, "name must be between 5 and 50 characters"),
		)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		presentation, err := queriesStore.CreatePresentation(ctx, *input.Name)
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

func DeletePresentationHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := queriesStore.DeletePresentation(ctx, ID)
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
