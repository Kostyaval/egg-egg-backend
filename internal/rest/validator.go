package rest

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func setupValidator() {
	// Get json|query|params key instead of struct name for validationErr.Field()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tagVal := fld.Tag.Get("json")

		if tagVal == "" {
			tagVal = fld.Tag.Get("query")
		}

		if tagVal == "" {
			tagVal = fld.Tag.Get("params")
		}

		name := strings.SplitN(tagVal, ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
}
