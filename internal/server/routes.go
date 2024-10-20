package server

import (
	"log/slog"
	"net/http"
)

func routes(logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /presentations", CreatePresentationHandler(logger))

	return mux
}
