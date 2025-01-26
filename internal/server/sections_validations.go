package server

import (
	"time"

	"github.com/PabloVarg/presentation-timer/internal/validation"
)

func ValidateSectionName(v validation.Validator, name *string) {
	v.Check(
		"name",
		name,
		validation.CheckPointerNotNil("name must be given"),
		validation.StringCheckNotEmpty("name can not be empty"),
		validation.StringCheckLength(5, 50, "name must be between 5 and 50 characters"),
	)
}

func ValidateDuration(v validation.Validator, duration *time.Duration) {
	v.Check(
		"duration",
		duration,
		validation.CheckPointerNotNil("duration must be given"),
		validation.DurationCheckPositive("duration can not be negative"),
		validation.DurationCheckMin("duration can not be less than 1 second", time.Second),
	)
}

func ValidatePosition(v validation.Validator, position *int16) {
	if position == nil {
		return
	}

	v.Check(
		"position",
		position,
		validation.IntCheckNatural("position can not be negative"),
	)
}

func ValidateMovement(v validation.Validator, move *int32) {
	v.Check(
		"move",
		move,
		validation.CheckPointerNotNil("move must be given"),
	)
}
