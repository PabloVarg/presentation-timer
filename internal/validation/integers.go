package validation

import "reflect"

func checkInt(value any) int64 {
	reflectValue := reflect.ValueOf(value)

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
