package internal

import (
	"todo-server/models"

	"github.com/go-playground/validator/v10"
)

const (
	ErrorCodeValidationFailed = "validation_failed"
	ErrorCodeErrorMessage     = "error_message"
)

func GetCustomErrorMessage(fe validator.FieldError) string {
	switch fe.Field() {
	case "Name":
		switch fe.Tag() {
		case "required":
			return "The Name field is required."
		case "min":
			return "The Name field must be at least 3 characters long."
		case "max":
			return "The name field must be less than 1000 characters long."
		}
	}

	return fe.Error()
}

func ConstructInvalidFieldData(error error) []models.InvalidField {
	var result []models.InvalidField

	for _, err := range error.(validator.ValidationErrors) {
		result = append(result, models.InvalidField{
			ErrorMessage: GetCustomErrorMessage(err),
			Field:        err.Field(),
			IsValid:      true,
		})
	}

	return result
}
