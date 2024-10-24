package validation

func checkInt(value any) int64 {
	switch value.(type) {
	case int8:
		return int64(value.(int8))
	case int16:
		return int64(value.(int16))
	case int32:
		return int64(value.(int32))
	case int64:
		return value.(int64)
	default:
		panic("validating wrong types")
	}
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
