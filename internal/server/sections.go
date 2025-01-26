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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		ValidateSectionName(v, input.Name)
		ValidateDuration(v, input.Duration)
		ValidatePosition(v, input.Position)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if input.Position == nil {
			position, err := queriesStore.MaxPosition(ctx, presentationID)
			if err != nil {
				helpers.InternalError(w, logger, err)
				return
			}

			input.Position = &position
		}

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

func UpdateSectionHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type input struct {
		Name     *string        `json:"name"`
		Duration *time.Duration `json:"duration"`
		Position *int16         `json:"position"`
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
		ValidateSectionName(v, input.Name)
		ValidateDuration(v, input.Duration)
		ValidatePosition(v, input.Position)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := queriesStore.UpdateSection(ctx, queries.UpdateSectionParams{
			ID:       ID,
			Name:     *input.Name,
			Duration: *input.Duration,
			Position: *input.Position,
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

func PatchSectionHandler(logger *slog.Logger, queriesStore *queries.Queries) http.Handler {
	type input struct {
		Name     *string        `json:"name"`
		Duration *time.Duration `json:"duration"`
		Position *int16         `json:"position"`
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
			ValidateSectionName(v, input.Name)
		}
		if input.Duration != nil {
			ValidateDuration(v, input.Duration)
		}
		if input.Position != nil {
			ValidatePosition(v, input.Position)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := queriesStore.PatchSection(ctx, queries.PatchSectionParams{
			ID:       ID,
			Name:     input.Name,
			Duration: input.Duration,
			Position: input.Position,
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

func MoveSectionHandler(
	logger *slog.Logger,
	queriesStore *queries.Queries,
) http.Handler {
	type input struct {
		Move *int32 `json:"move"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, v := helpers.ParseID(r, "id")
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		var input input
		err := helpers.ReadJSON(r.Body, &input)
		if err != nil {
			helpers.BadRequest(w, err.Error())
			return
		}

		v = validation.New()
		ValidateMovement(v, input.Move)
		if !v.Valid() {
			helpers.UnprocessableContent(w, v.Errors())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = queriesStore.CleanPositionsBySectionGroup(ctx, ID)
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}

		err = queriesStore.MoveSection(ctx, queries.MoveSectionParams{
			ID:      ID,
			Column2: *input.Move,
		})
		if err != nil {
			helpers.InternalError(w, logger, err)
			return
		}
	})
}
