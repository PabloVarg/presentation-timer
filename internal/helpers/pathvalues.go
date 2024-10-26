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
		validation.StringCheckNotEmpty(fmt.Sprintf("%s can not be empty", key)),
	)
	if !v.Valid() {
		return 0, v
	}

	ID, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		v.AddErrors(key, fmt.Sprintf("%d is not an integer", ID))
		return 0, v
	}

	v.Check(
		key,
		ID,
		validation.IntCheckPositive(fmt.Sprintf("%s is not positive", key)),
	)

	return ID, v
}
