package server

import "github.com/PabloVarg/presentation-timer/internal/validation"

func ValidateName(v validation.Validator, name *string) {
	v.Check(
		"name",
		name,
		validation.CheckPointerNotNil("name must be given"),
		validation.StringCheckNotEmpty("name can't be empty"),
		validation.StringCheckLength(5, 50, "name must be between 5 and 50 characters"),
	)
}
