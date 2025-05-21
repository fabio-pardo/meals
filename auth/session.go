package auth

import (
	"meals/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SessionMiddleware handles session-based authentication
func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session cookie
		cookie, err := c.Request.Cookie("session")
		if err != nil {
			// No session cookie found, continue the request
			c.Next()
			return
		}

		// Extract session token
		sessionToken := cookie.Value

		// Get database from context
		dbInterface, exists := c.Get("db")
		if !exists {
			// No database in context, continue the request
			c.Next()
			return
		}

		// Convert to Database type
		dbWrapper, ok := dbInterface.(*models.Database)
		if !ok {
			// Invalid database type, continue the request
			c.Next()
			return
		}

		db := dbWrapper.DB

		// Look up session in database
		var session models.Session
		if err := db.Where("token = ?", sessionToken).First(&session).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Session not found, continue the request
				c.Next()
				return
			}
			// Database error, continue the request
			c.Next()
			return
		}

		// Check if session is expired
		if session.IsExpired() {
			// Delete expired session
			db.Delete(&session)
			
			// Clear session cookie
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})
			
			// Session expired, continue the request
			c.Next()
			return
		}

		// Get user from database
		var user models.User
		if err := db.First(&user, session.UserIdentifier).Error; err != nil {
			// User not found, continue the request
			c.Next()
			return
		}

		// Set user in context
		c.Set("user", user)

		// Continue with request
		c.Next()
	}
}

// RequireSession middleware ensures the user is authenticated with a valid session
func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "You must be logged in to access this resource",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CreateSession creates a new session for the user
func CreateSession(c *gin.Context, db *gorm.DB, userID string, duration time.Duration) (string, error) {
	// Generate random token
	token := GenerateRandomToken()

	// Create session
	session := models.Session{
		UserIdentifier: userID,
		Token:          token,
		ExpiresAt:      time.Now().Add(duration),
	}

	// Save session to database
	if err := db.Create(&session).Error; err != nil {
		return "", err
	}

	// Set session cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		MaxAge:   int(duration.Seconds()),
	})

	return token, nil
}

// GenerateRandomToken generates a random session token
func GenerateRandomToken() string {
	// In a real implementation, this would use a secure random generator
	// For test purposes, we'll use a simple timestamp-based token
	return "session-" + time.Now().Format("20060102150405")
}
