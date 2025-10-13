package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AppError represents a structured application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeBadRequest          = "bad_request"
	ErrCodeUnauthorized        = "unauthorized"
	ErrCodeForbidden           = "forbidden"
	ErrCodeNotFound            = "not_found"
	ErrCodeConflict            = "conflict"
	ErrCodeValidation          = "validation_error"
	ErrCodeInternal            = "internal_error"
	ErrCodeDatabase            = "database_error"
	ErrCodeInvalidInput        = "invalid_input"
	ErrCodeResourceNotFound    = "resource_not_found"
	ErrCodeDuplicateResource   = "duplicate_resource"
	ErrCodeInsufficientBalance = "insufficient_balance"
	ErrCodePaymentFailed       = "payment_failed"
	ErrCodeOrderNotFound       = "order_not_found"
	ErrCodeInvalidStatus       = "invalid_status"
	ErrCodeUnprocessable       = "unprocessable_entity"
)

// Predefined errors
var (
	ErrBadRequest          = NewBadRequest("Invalid request")
	ErrUnauthorized        = NewUnauthorized("Authentication required")
	ErrForbidden           = NewForbidden("Access denied")
	ErrNotFound            = NewNotFound("Resource not found")
	ErrInternalServer      = NewInternal("Internal server error")
	ErrInvalidInput        = NewValidation("Invalid input provided")
	ErrDatabaseError       = NewDatabase("Database operation failed")
	ErrResourceNotFound    = NewNotFound("Requested resource not found")
	ErrDuplicateResource   = NewConflict("Resource already exists")
	ErrInsufficientBalance = NewBadRequest("Insufficient balance")
	ErrPaymentFailed       = NewBadRequest("Payment processing failed")
	ErrOrderNotFound       = NewNotFound("Order not found")
	ErrInvalidStatus       = NewBadRequest("Invalid status transition")
)

// Constructor functions
func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewForbidden(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewNotFound(message string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

func NewConflict(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewValidation(message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewInternal(message string) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

func NewDatabase(message string) *AppError {
	return &AppError{
		Code:       ErrCodeDatabase,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

func NewUnprocessable(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnprocessable,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// HandleError processes an error and sends appropriate response
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, appErr)
		return
	}

	// Handle GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, NewNotFound("Resource not found"))
		return
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, NewConflict("Resource already exists"))
		return
	}

	// Handle validation errors from Gin
	if _, ok := err.(gin.Error); ok {
		c.JSON(http.StatusBadRequest, NewValidation(err.Error()))
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, NewInternal("An unexpected error occurred").WithDetails(err.Error()))
}

// RespondError is a helper to respond with an AppError
func RespondError(c *gin.Context, err *AppError) {
	c.JSON(err.StatusCode, err)
}

// RespondSuccess is a helper to respond with success
func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// RespondCreated is a helper for 201 Created responses
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// RespondOK is a helper for 200 OK responses
func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondNoContent is a helper for 204 No Content responses
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// NewValidationErrors creates a new validation errors response
func NewValidationErrors(errors []ValidationError) *ValidationErrors {
	return &ValidationErrors{
		Code:    ErrCodeValidation,
		Message: "Validation failed",
		Errors:  errors,
	}
}

// RespondValidationErrors sends validation errors response
func RespondValidationErrors(c *gin.Context, errors []ValidationError) {
	c.JSON(http.StatusBadRequest, NewValidationErrors(errors))
}
