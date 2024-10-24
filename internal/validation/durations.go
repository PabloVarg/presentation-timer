package validation

import "time"

func checkDuration(value any) time.Duration {
	result, ok := value.(time.Duration)
	if !ok {
		panic("validating wrong types")
	}

	return result
}

func DurationCheckPositive(message string) ValidationFunc {
	return func(value any) (bool, string) {
		durationValue := checkDuration(value)

		return durationValue >= 0, message
	}
}

func DurationCheckMin(message string, minDuration time.Duration) ValidationFunc {
	return func(value any) (bool, string) {
		durationValue := checkDuration(value)

		return durationValue >= minDuration, message
	}
}
