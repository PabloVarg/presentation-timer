package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	res, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)
	return nil
}

func ReadJSON(r io.Reader, data any) error {
	if err := json.NewDecoder(r).Decode(data); err != nil {
		return fmt.Errorf("malformed input (%w)", err)
	}
	return nil
}
