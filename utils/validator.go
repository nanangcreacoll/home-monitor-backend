package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

func ValidationError(err error) map[string]string {
	if err == nil {
		return nil
	}

	errorsMap := make(map[string]string)

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		errorsMap["error"] = "Invalid request"
		return errorsMap
	}

	for _, fieldError := range ve {
		field := fieldError.Field()
		switch fieldError.Tag() {
		case "required":
			errorsMap[field] = field + " is required"
		case "min":
			errorsMap[field] = field + " must be at least " + fieldError.Param() + " characters"
		case "max":
			errorsMap[field] = field + " must be at most " + fieldError.Param() + " characters"
		default:
			errorsMap[field] = "Invalid value for " + field
		}
	}

	return errorsMap
}
