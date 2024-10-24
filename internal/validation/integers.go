package validation

func checkInt(value any) int {
	result, ok := value.(int)
	if !ok {
		panic("validating wrong types")
	}

	return result
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
