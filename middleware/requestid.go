package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header key for the request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for the request ID
	RequestIDContextKey = "requestID"
)

// RequestID middleware generates a unique request ID for each request
// If a request ID is already present in the incoming request headers, it will use that instead
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in the request headers
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			// Generate a new request ID if none is present
			requestID = uuid.New().String()
		}

		// Set the request ID in the context for use by other middleware and handlers
		c.Set(RequestIDContextKey, requestID)

		// Add request ID to response headers
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID returns the request ID from the gin context
// Returns empty string if request ID is not found
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDContextKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
