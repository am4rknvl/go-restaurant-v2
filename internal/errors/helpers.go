package errors

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// HandleBindError handles JSON binding errors
func HandleBindError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's a validation error
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		var fieldErrors []ValidationError
		for _, fieldErr := range validationErrs {
			fieldErrors = append(fieldErrors, ValidationError{
				Field:   fieldErr.Field(),
				Message: getValidationMessage(fieldErr),
			})
		}
		RespondValidationErrors(c, fieldErrors)
		return
	}

	// Generic binding error
	RespondError(c, NewValidation("Invalid request body").WithDetails(err.Error()))
}

// getValidationMessage returns a user-friendly validation message
func getValidationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldErr.Field())
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fieldErr.Field(), fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fieldErr.Field(), fieldErr.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldErr.Field(), fieldErr.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldErr.Field(), fieldErr.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldErr.Field())
	default:
		return fmt.Sprintf("%s is invalid", fieldErr.Field())
	}
}

// HandleDatabaseError handles database-related errors
func HandleDatabaseError(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	appErr := NewDatabase(fmt.Sprintf("Failed to %s", operation)).WithDetails(err.Error())
	RespondError(c, appErr)
}

// HandleNotFoundError handles resource not found errors
func HandleNotFoundError(c *gin.Context, resourceType string) {
	appErr := NewNotFound(fmt.Sprintf("%s not found", resourceType))
	RespondError(c, appErr)
}

// HandleConflictError handles resource conflict errors
func HandleConflictError(c *gin.Context, message string) {
	appErr := NewConflict(message)
	RespondError(c, appErr)
}

// HandleUnauthorizedError handles unauthorized access errors
func HandleUnauthorizedError(c *gin.Context, message string) {
	if message == "" {
		message = "Authentication required"
	}
	appErr := NewUnauthorized(message)
	RespondError(c, appErr)
}

// HandleForbiddenError handles forbidden access errors
func HandleForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = "Access denied"
	}
	appErr := NewForbidden(message)
	RespondError(c, appErr)
}

// HandleInternalError handles internal server errors
func HandleInternalError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	appErr := NewInternal("An unexpected error occurred").WithDetails(err.Error())
	RespondError(c, appErr)
}

// ValidateRequired checks if a required field is present
func ValidateRequired(c *gin.Context, fieldName string, value interface{}) bool {
	if value == nil || value == "" {
		RespondError(c, NewValidation(fmt.Sprintf("%s is required", fieldName)))
		return false
	}
	return true
}

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(c *gin.Context, fieldName, value string) bool {
	if value == "" {
		RespondError(c, NewValidation(fmt.Sprintf("%s must be a valid UUID", fieldName)))
		return false
	}
	// Basic UUID validation (can be enhanced with regex)
	if len(value) != 36 {
		RespondError(c, NewValidation(fmt.Sprintf("%s must be a valid UUID", fieldName)))
		return false
	}
	return true
}
