package handlers

import (
	"meals/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status    int    `json:"-"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Error implements error.
func (e ErrorResponse) Error() string {
	panic("unimplemented")
}

// Common error codes
const (
	ErrBadRequest        = "BAD_REQUEST"
	ErrNotFound          = "NOT_FOUND"
	ErrInternalServer    = "INTERNAL_SERVER_ERROR"
	ErrUnauthorized      = "UNAUTHORIZED"
	ErrForbidden         = "FORBIDDEN"
	ErrValidation        = "VALIDATION_ERROR"
	ErrDatabaseOperation = "DATABASE_ERROR"
	ErrResourceExists    = "RESOURCE_EXISTS"
	ErrRelationship      = "RELATIONSHIP_ERROR"
)

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, err ErrorResponse) {
	if err.Status == 0 {
		err.Status = http.StatusInternalServerError
	}

	// Get request ID from context and add it to the response
	requestID := middleware.GetRequestID(c)

	c.JSON(err.Status, gin.H{
		"error": gin.H{
			"code":       err.Code,
			"message":    err.Message,
			"details":    err.Details,
			"request_id": requestID,
		},
	})
}

// NotFoundError returns a standardized not found error
func NotFoundError(resource string) ErrorResponse {
	return ErrorResponse{
		Status:  http.StatusNotFound,
		Code:    ErrNotFound,
		Message: resource + " not found",
	}
}

// BadRequestError returns a standardized bad request error
func BadRequestError(message string) ErrorResponse {
	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    ErrBadRequest,
		Message: message,
	}
}

// ValidationError returns a standardized validation error
func ValidationError(message string, details any) ErrorResponse {
	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    ErrValidation,
		Message: message,
		Details: details,
	}
}

// DatabaseError returns a standardized database error
func DatabaseError(message string) ErrorResponse {
	return ErrorResponse{
		Status:  http.StatusInternalServerError,
		Code:    ErrDatabaseOperation,
		Message: message,
	}
}

// RelationshipError returns a standardized relationship error for related entities
func RelationshipError(message string, details any) ErrorResponse {
	return ErrorResponse{
		Status:  http.StatusBadRequest,
		Code:    ErrRelationship,
		Message: message,
		Details: details,
	}
}

// Custom error types for transaction-compatible errors

// AppError is an interface for all application errors
type AppError interface {
	Error() string
	ToResponse() ErrorResponse
}

// ValidationErrorType represents validation errors
type ValidationErrorType struct {
	Message string
	Details any
}

func (e ValidationErrorType) Error() string {
	return e.Message
}

func (e ValidationErrorType) ToResponse() ErrorResponse {
	return ValidationError(e.Message, e.Details)
}

// NotFoundErrorType represents not found errors
type NotFoundErrorType struct {
	Resource string
}

func (e NotFoundErrorType) Error() string {
	return e.Resource + " not found"
}

func (e NotFoundErrorType) ToResponse() ErrorResponse {
	return NotFoundError(e.Resource)
}

// RelationshipErrorType represents relationship errors between entities
type RelationshipErrorType struct {
	Message string
	Details any
}

func (e RelationshipErrorType) Error() string {
	return e.Message
}

func (e RelationshipErrorType) ToResponse() ErrorResponse {
	return RelationshipError(e.Message, e.Details)
}

// DatabaseErrorType represents database operation errors
type DatabaseErrorType struct {
	Message string
	Details any
}

func (e DatabaseErrorType) Error() string {
	return e.Message
}

func (e DatabaseErrorType) ToResponse() ErrorResponse {
	return DatabaseError(e.Message)
}

// UnauthorizedErrorType represents unauthorized access errors
type UnauthorizedErrorType struct {
	Message string
}

func (e UnauthorizedErrorType) Error() string {
	return e.Message
}

func (e UnauthorizedErrorType) ToResponse() ErrorResponse {
	return ErrorResponse{
		Status:  http.StatusUnauthorized,
		Code:    ErrUnauthorized,
		Message: e.Message,
	}
}

// BadRequestErrorType represents bad request errors
type BadRequestErrorType struct {
	Message string
}

func (e BadRequestErrorType) Error() string {
	return e.Message
}

func (e BadRequestErrorType) ToResponse() ErrorResponse {
	return BadRequestError(e.Message)
}

// HandleAppError handles all application errors in a uniform way
func HandleAppError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if appErr, ok := err.(AppError); ok {
		// If it's an application error, convert to response
		RespondWithError(c, appErr.ToResponse())
		return true
	}

	// Handle gorm specific errors
	if err == gorm.ErrRecordNotFound {
		RespondWithError(c, NotFoundError("Resource"))
		return true
	}

	// Default case
	RespondWithError(c, DatabaseError(err.Error()))
	return true
}
