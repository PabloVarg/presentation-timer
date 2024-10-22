package server

import (
	"log/slog"
	"net/http"

	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
)

func routes(logger *slog.Logger, queries *queries.Queries) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /presentations", ListPresentationsHandler(logger, queries))
	mux.Handle("GET /presentations/{id}", ListPresentationHandler(logger, queries))
	mux.Handle("POST /presentations", CreatePresentationHandler(logger, queries))
	mux.Handle("DELETE /presentations/{id}", DeletePresentationHandler(logger, queries))

	return mux
}
