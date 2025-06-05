package internal

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	return validator.New()
}

func ValidateStruct(v *validator.Validate, s interface{}) error {
	err := v.Struct(s)
	if err != nil {
		return err
	}
	return nil
}
func RegisterCustomValidations(validate *validator.Validate) {
	validate.RegisterValidation("Alphanumericspaceunderhyphen", Alphanumericspaceunderhyphen)
}

func Alphanumericspaceunderhyphen(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return false
	}
	allowChars := regexp.MustCompile(`^[A-Za-z0-9\-_ ]+$`)
	return allowChars.MatchString(str)
}
