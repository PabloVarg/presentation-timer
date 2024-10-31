package server

import (
	"log/slog"
	"net/http"

	"github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
)

func routes(logger *slog.Logger, queries *queries.Queries) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /presentations", ListPresentationsHandler(logger, queries))
	mux.Handle("GET /presentations/{id}", GetPresentationHandler(logger, queries))
	mux.Handle("POST /presentations", CreatePresentationHandler(logger, queries))
	mux.Handle("PUT /presentations/{id}", PutPresentationHandler(logger, queries))
	mux.Handle("DELETE /presentations/{id}", DeletePresentationHandler(logger, queries))

	mux.Handle("GET /presentations/{presentation_id}/sections", ListSectionsHandler(logger, queries))
	mux.Handle("POST /presentations/{presentation_id}/sections", CreateSectionHandler(logger, queries))

	mux.Handle("GET /sections/{id}", GetSectionHandler(logger, queries))
	mux.Handle("DELETE /sections/{id}", DeleteSectionHandler(logger, queries))

	return mux
}
