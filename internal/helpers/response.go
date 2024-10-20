package helpers

import "net/http"

type ErrorResponse struct {
	error string
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusBadRequest, ErrorResponse{
		error: message,
	})
}
