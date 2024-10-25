package validation

import (
	"reflect"
	"unicode/utf8"
)

func checkString(value any) string {
	reflectValue := extractValue(value)
	if extractValue(value).Kind() != reflect.String {
		return ""
	}

	return reflectValue.String()
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
