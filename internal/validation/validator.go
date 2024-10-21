package validation

type Validator struct {
	errors map[string][]string
}

type ValidationFunc = func(value any) (bool, string)

func New() Validator {
	return Validator{
		errors: map[string][]string{},
	}
}

func (v Validator) Errors() map[string][]string {
	return v.errors
}

func (v Validator) Valid() bool {
	return len(v.errors) == 0
}

func (v Validator) AddErrors(key string, messages ...string) {
	if _, ok := v.errors[key]; ok {
		v.errors[key] = append(v.errors[key], messages...)
		return
	}

	v.SetErrors(key, messages...)
}

func (v Validator) SetErrors(key string, messages ...string) {
	v.errors[key] = messages
}

func (v Validator) Check(key string, value any, validations ...ValidationFunc) {
	validationErrors := make([]string, 0)

	for _, validation := range validations {
		ok, message := validation(value)
		if ok {
			continue
		}

		validationErrors = append(validationErrors, message)
	}

	v.AddErrors(key, validationErrors...)
}
