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
	"github.com/jackc/pgx/v5/pgconn"
)

const SectionsPageSize = 20

var SectionsSortFields = []string{"name", "duration", "position"}

func ListSectionsHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type output struct {
		Data     []queries.Section `json:"data"`
		PageInfo filters.PageInfo  `json:"page_info"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		presentationID, v := helpers.ParseID(r, "presentation_id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		f, v := filters.FromRequest(r, SectionsPageSize, SectionsSortFields...)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		sections, err := queriesStore.GetSections(ctx, queries.GetSectionsParams{
			PresentationID: presentationID,
			Direction:      f.QuerySortDirection(),
			SortBy:         f.QuerySortBy(),
			QueryLimit:     f.QueryLimit(),
			QueryOffset:    f.QueryOffset(),
		})
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		totalRows, err := queriesStore.GetSectionsMetadata(ctx, presentationID)
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		if err := helpers.WriteJSON(w, http.StatusOK, output{
			Data:     sections,
			PageInfo: f.PageInfo(totalRows),
		}); err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}

func GetSectionHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		section, err := queriesStore.GetSection(ctx, ID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.NotFound(w, r)
			default:
				helpers.InternalError(w, logger, err)
			}
			return
		}

		if err := helpers.WriteJSON(w, http.StatusOK, section); err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}

func CreateSectionHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type input struct {
		Name     *string        `json:"name"`
		Duration *time.Duration `json:"duration"`
		Position *int16         `json:"position"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input input

		presentationID, v := helpers.ParseID(r, "presentation_id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		if err := helpers.ReadJSON(r.Body, &input); err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		v = validation.New()
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
			Presentation: presentationID,
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
