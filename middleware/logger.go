package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware logs request information with request IDs
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Get request ID (will be empty string if not set)
		requestID := GetRequestID(c)
		requestIDField := ""
		if requestID != "" {
			requestIDField = "[" + requestID + "] "
		}

		// Calculate latency
		latency := time.Since(startTime)

		// Log request details
		log.Printf("%s%s %s %s | %d | %v",
			requestIDField,
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Writer.Status(),
			latency,
		)
	}
}
