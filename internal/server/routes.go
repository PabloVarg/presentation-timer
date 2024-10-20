package server

import (
	"log/slog"
	"net/http"

	queries "github.com/PabloVarg/presentation-timer/internal/queries/sqlc"
)

func routes(logger *slog.Logger, queries *queries.Queries) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /presentations", CreatePresentationHandler(logger, queries))

	return mux
}
