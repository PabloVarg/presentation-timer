package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

func ReadJSON(r io.Reader, data any) error {
	if err := json.NewDecoder(r).Decode(data); err != nil {
		return fmt.Errorf("malformed input (%w)", err)
	}
	return nil
}
