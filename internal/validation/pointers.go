package validation

import (
	"reflect"
)

func CheckPointerNotNil(message string) ValidationFunc {
	return func(value any) (bool, string) {
		reflectValue := reflect.ValueOf(value)

		return !reflectValue.IsNil(), message
	}
}

func extractValue(value any) reflect.Value {
	reflectValue := reflect.ValueOf(value)

	if reflectValue.Kind() == reflect.Pointer && !reflectValue.IsNil() {
		return reflectValue.Elem()
	}
	if reflectValue.Kind() == reflect.Pointer && reflectValue.IsNil() {
		return reflect.Zero(reflectValue.Type().Elem())
	}

	return reflectValue
}
