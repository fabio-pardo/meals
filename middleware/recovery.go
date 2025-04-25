package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery middleware recovers from any panics and logs them with request IDs
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get the request ID
				requestID := GetRequestID(c)
				requestIDField := ""
				if requestID != "" {
					requestIDField = "[" + requestID + "] "
				}

				// Log the error and stack trace
				debugStack := debug.Stack()
				log.Printf("%sPANIC RECOVERED: %v\n%s", requestIDField, err, string(debugStack))

				// Send a 500 response with the request ID
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":       "INTERNAL_SERVER_ERROR",
						"message":    "An unexpected error occurred",
						"request_id": requestID,
					},
				})

				// Abort the request
				c.Abort()
			}
		}()
		c.Next()
	}
}
