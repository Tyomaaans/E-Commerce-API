package pkg

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var instance *validator.Validate

func New() *validator.Validate {
	instance = validator.New()
	registerCustomValidators(instance)
	return instance
}

func registerCustomValidators(v *validator.Validate) {
	v.RegisterValidation("alphaspaceunicode", alphaSpaceUnicode)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("phone", validatePhone)
}

func validatePassword(fl validator.FieldLevel) bool {
	s := fl.Field().String()

	var hasUpper, hasNumber, hasSymbol bool

	for _, c := range s {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasNumber = true
		case !unicode.IsLetter(c) && !unicode.IsDigit(c) && !unicode.IsSpace(c):
			hasSymbol = true
		case unicode.IsSpace(c):
			return false
		}
	}

	return hasUpper && hasNumber && hasSymbol
}

func alphaSpaceUnicode(fl validator.FieldLevel) bool {
	for _, r := range fl.Field().String() {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	re := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	return re.MatchString(phone)
}