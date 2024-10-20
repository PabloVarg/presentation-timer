package server

import (
	"log/slog"
	"net/http"
)

func routes(_ *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}
