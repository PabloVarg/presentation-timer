package validation

func checkInt(value any) int64 {
	reflectValue := extractValue(value)
	if reflectValue.IsZero() {
		return 0
	}

	return reflectValue.Int()
}

func IntCheckNatural(message string) ValidationFunc {
	return func(value any) (bool, string) {
		intValue := checkInt(value)

		return intValue >= 0, message
	}
}

func IntCheckPositive(message string) ValidationFunc {
	return func(value any) (bool, string) {
		intValue := checkInt(value)

		return intValue > 0, message
	}
}

func IntCheckMax(maxValue int64, message string) ValidationFunc {
	return func(value any) (bool, string) {
		intValue := checkInt(value)

		return intValue <= maxValue, message
	}
}
