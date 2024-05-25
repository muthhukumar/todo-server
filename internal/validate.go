package internal

import (
	"todo-server/models"
)

func ValidateTodo(task models.Task) (isValid bool, validationResult []models.FieldValidation) {
	if task.Name == "" {
		validationResult = append(validationResult, models.FieldValidation{
			Field:        "name",
			IsValid:      false,
			ErrorMessage: "Please enter valid task",
		})
		isValid = false

		return
	}

	isValid = true

	return
}
