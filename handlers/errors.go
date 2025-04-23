package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  int    `json:"-"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Common error codes
const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrNotFound            = "NOT_FOUND"
	ErrInternalServer      = "INTERNAL_SERVER_ERROR"
	ErrUnauthorized        = "UNAUTHORIZED"
	ErrForbidden           = "FORBIDDEN"
	ErrValidation          = "VALIDATION_ERROR"
	ErrDatabaseOperation   = "DATABASE_ERROR"
	ErrResourceExists      = "RESOURCE_EXISTS"
)

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, err ErrorResponse) {
	if err.Status == 0 {
		err.Status = http.StatusInternalServerError
	}
	
	c.JSON(err.Status, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
			"details": err.Details,
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
