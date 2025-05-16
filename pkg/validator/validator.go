package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ParseError(err error) map[string]string {
	errors := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			errors[fe.Field()] = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", fe.Field(), fe.Tag())
		}
	} else if err != nil { // Non-validator errors
		errors["error"] = err.Error()
	}
	return errors
}
