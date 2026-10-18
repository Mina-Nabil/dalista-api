package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors turns a validator error into a flat list of
// human-readable messages for the ValidationErrorResponse envelope.
func FormatValidationErrors(err error) []string {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}
	}
	msgs := make([]string, 0, len(validationErrs))
	for _, fe := range validationErrs {
		msgs = append(msgs, formatValidationError(fe))
	}
	return msgs
}

func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "e164":
		return fmt.Sprintf("%s must be a valid phone number in E.164 format", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
