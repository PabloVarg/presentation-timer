package helpers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PabloVarg/presentation-timer/internal/validation"
)

func ParseID(r *http.Request, key string) (int64, validation.Validator) {
	v := validation.New()

	strID := r.PathValue(key)
	v.Check(
		key,
		strID,
		validation.CheckNotEmpty(fmt.Sprintf("%s can not be empty", key)),
	)
	if !v.Valid() {
		return 0, v
	}

	ID, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		v.AddErrors(key, fmt.Sprintf("%d is not an integer", ID))
		return 0, v
	}

	return ID, v
}
