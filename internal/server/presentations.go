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
	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
	"github.com/PabloVarg/presentation-timer/internal/validation"
)

const PresentationsPageSize = 20

var PresentationsSortFields = []string{"name"}

func ListPresentationsHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type output struct {
		Data     []queries.Presentation `json:"data"`
		PageInfo filters.PageInfo       `json:"page_info"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, v := filters.FromRequest(r, PresentationsPageSize, PresentationsSortFields...)
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

		totalRows, err := queriesStore.GetPresentationsMetadata(ctx)
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		if err := helpers.WriteJSON(w, http.StatusOK, output{
			Data:     presentations,
			PageInfo: f.PageInfo(totalRows),
		}); err != nil {
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
		ValidatePresentationName(v, input.Name)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func PutPresentationHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type Input struct {
		Name *string `json:"name"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input Input

		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		if err := helpers.ReadJSON(r.Body, &input); err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		v = validation.New()
		ValidatePresentationName(v, input.Name)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := queriesStore.UpdatePresentation(ctx, queries.UpdatePresentationParams{
			ID:   ID,
			Name: *input.Name,
		})
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
		if rows == 0 {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func PatchPresentationHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type input struct {
		Name *string `json:"name"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input input

		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		if err := helpers.ReadJSON(r.Body, &input); err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		v = validation.New()
		if input.Name != nil {
			ValidatePresentationName(v, input.Name)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := queriesStore.PatchPresentation(ctx, queries.PatchPresentationParams{
			ID:   ID,
			Name: input.Name,
		})
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
		if rows == 0 {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func DeletePresentationHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
