package validation

import "unicode/utf8"

func checkString(value any) string {
	result, ok := value.(string)
	if !ok {
		panic("validating wrong types")
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
		runeCount := utf8.RuneCountInString(strValue)

		return minLength <= runeCount && runeCount <= maxLength, message
	}
}

func CheckMinLen(minLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return utf8.RuneCountInString(strValue) >= minLength, message
	}
}

func CheckMaxLen(maxLength int, message string) ValidationFunc {
	return func(value any) (bool, string) {
		strValue := checkString(value)

		return utf8.RuneCountInString(strValue) <= maxLength, message
	}
}
