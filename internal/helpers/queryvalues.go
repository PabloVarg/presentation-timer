package helpers

import (
	"net/http"
	"strconv"
)

func QueryInt32(r *http.Request, key string, defaultValue int32) (int32, error) {
	if r.URL.Query().Get(key) == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.ParseInt(r.URL.Query().Get(key), 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(intValue), nil
}
