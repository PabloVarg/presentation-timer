package validation

import "unicode/utf8"

func checkString(value any) string {
	result, ok := value.(string)
	if !ok {
		panic("validating wrong types")
	}

	return result
}

func StringCheckNotEmpty(message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return strValue != "", message
	}
}

func StringCheckEmpty(message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return strValue == "", message
	}
}

func StringCheckLength(minLength, maxLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)
		runeCount := utf8.RuneCountInString(strValue)

		return minLength <= runeCount && runeCount <= maxLength, message
	}
}

func StringCheckMinLen(minLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return utf8.RuneCountInString(strValue) >= minLength, message
	}
}

func StringCheckMaxLen(maxLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return utf8.RuneCountInString(strValue) <= maxLength, message
	}
}
