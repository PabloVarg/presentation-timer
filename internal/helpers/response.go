package helpers

import (
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusBadRequest, ErrorResponse{
		Error: message,
	})
}

func InternalError(w http.ResponseWriter, logger *slog.Logger, err error) {
	logger.Error("server error", "err", err)
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: "an error has ocurred",
	})
}
