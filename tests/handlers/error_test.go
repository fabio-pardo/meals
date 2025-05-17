package handlers_test

import (
	"encoding/json"
	"meals/handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ErrorResponse represents the JSON structure of error responses
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestErrorHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ValidationError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create a validation error
		err := handlers.ValidationErrorType{Message: "Test validation error", Details: "test"}

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled
		assert.True(t, handled, "Expected error to be handled")
		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code, "Expected VALIDATION_ERROR type")
		assert.Equal(t, "Test validation error", response.Error.Message, "Expected correct error message")
	})

	t.Run("NotFoundError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create a not found error
		err := handlers.NotFoundErrorType{Resource: "User"}

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled
		assert.True(t, handled, "Expected error to be handled")
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404 Not Found status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "NOT_FOUND", response.Error.Code, "Expected NOT_FOUND type")
		assert.Equal(t, "User not found", response.Error.Message, "Expected correct error message")
	})

	t.Run("DatabaseError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create a database error
		err := handlers.DatabaseErrorType{Message: "Failed to connect to database"}

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled
		assert.True(t, handled, "Expected error to be handled")
		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected 500 Internal Server Error status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "DATABASE_ERROR", response.Error.Code, "Expected DATABASE_ERROR type")
		assert.Equal(t, "Failed to connect to database", response.Error.Message, "Expected correct error message")
	})

	t.Run("UnauthorizedError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create an unauthorized error
		err := handlers.UnauthorizedErrorType{Message: "Access denied"}

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled
		assert.True(t, handled, "Expected error to be handled")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401 Unauthorized status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "UNAUTHORIZED", response.Error.Code, "Expected UNAUTHORIZED type")
		assert.Equal(t, "Access denied", response.Error.Message, "Expected correct error message")
	})

	t.Run("BadRequestError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create a bad request error
		err := handlers.BadRequestErrorType{Message: "Invalid request parameters"}

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled
		assert.True(t, handled, "Expected error to be handled")
		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "BAD_REQUEST", response.Error.Code, "Expected BAD_REQUEST type")
		assert.Equal(t, "Invalid request parameters", response.Error.Message, "Expected correct error message")
	})

	t.Run("RelationshipError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create a relationship error using the RelationshipErrorType
		err := handlers.RelationshipErrorType{Message: "Cannot delete meal referenced by menu"}

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled
		assert.True(t, handled, "Expected error to be handled")
		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected 400 Bad Request status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "RELATIONSHIP_ERROR", response.Error.Code, "Expected RELATIONSHIP_ERROR type")
		assert.Equal(t, "Cannot delete meal referenced by menu", response.Error.Message, "Expected correct error message")
	})

	t.Run("NonAppError", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Create a standard error that doesn't implement the AppError interface
		err := assert.AnError // Use the built-in error from testify

		// Handle the error
		handled := handlers.HandleAppError(c, err)

		// Verify error was handled with database error type
		assert.True(t, handled, "Expected standard error to be handled by HandleAppError")
		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected 500 Internal Server Error status code")

		// Parse response
		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "DATABASE_ERROR", response.Error.Code, "Expected DATABASE_ERROR type")
		assert.Equal(t, "assert.AnError general error for testing", response.Error.Message, "Expected error message to be passed through")
	})
}
