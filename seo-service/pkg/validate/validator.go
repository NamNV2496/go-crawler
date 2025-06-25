package validate

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type IValidator interface {
	Validate(i interface{}) error
}
type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	validate := validator.New()
	accessTags := []string{
		"json",
		"param",
		"query",
		"header",
	}
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		for _, tag := range accessTags {
			name := strings.SplitN(fld.Tag.Get(tag), ",", 2)[0]
			if name == "-" {
				return ""
			}
		}
		return ""
	})
	validate.RegisterValidation("url-checker", func(fl validator.FieldLevel) bool {
		fields := fl.Field().Interface().([]string)
		for _, field := range fields {
			err := validate.Var(field, "url")
			if err != nil {
				return false
			}
		}
		return true
	})
	return &Validator{
		validator: validate,
	}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
