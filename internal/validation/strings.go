package validation

func checkString(value any) string {
	result, ok := value.(string)
	if !ok {
		panic("validating wrong value types")
	}

	return result
}

func CheckNotEmpty(message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return strValue != "", message
	}
}

func CheckEmpty(message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return strValue == "", message
	}
}

func CheckLength(minLength, maxLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return minLength <= len(strValue) && len(strValue) <= maxLength, message
	}
}

func CheckMinLen(minLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return len(strValue) >= minLength, message
	}
}

func CheckMaxLen(maxLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return len(strValue) <= maxLength, message
	}
}
