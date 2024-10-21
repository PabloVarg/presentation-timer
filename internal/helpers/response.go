package helpers

import (
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type UnprocessableErrorResponse struct {
	ErrorResponse
	Messages map[string][]string `json:"messages"`
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusBadRequest, ErrorResponse{
		Error: message,
	})
}

func UnprocessableContent(w http.ResponseWriter, messages map[string][]string) {
	WriteJSON(w, http.StatusUnprocessableEntity, UnprocessableErrorResponse{
		ErrorResponse: ErrorResponse{
			Error: "content is not valid",
		},
		Messages: messages,
	})
}

func InternalError(w http.ResponseWriter, logger *slog.Logger, err error) {
	logger.Error("server error", "err", err)
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: "an error has ocurred",
	})
}
