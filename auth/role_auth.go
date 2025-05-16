package auth

import (
	"meals/models"
	"meals/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse for auth middleware
type ErrorResponse struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
)

// sendError sends a standardized error response
func sendError(c *gin.Context, errResp ErrorResponse) {
	c.JSON(errResp.Status, gin.H{
		"error": gin.H{
			"code":    errResp.Code,
			"message": errResp.Message,
		},
	})
}

// RequireRole middleware validates that a user has the required role
func RequireRole(roles ...models.UserType) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		user, err := GetSessionUser(c.Request)
		if err != nil {
			sendError(c, ErrorResponse{
				Status:  http.StatusUnauthorized,
				Code:    ErrUnauthorized,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		// Set user in context for downstream handlers
		c.Set("user", user)

		// Get the full user model with user type from database
		var dbUser models.User
		if err := store.DB.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
			sendError(c, ErrorResponse{
				Status:  http.StatusUnauthorized,
				Code:    ErrUnauthorized,
				Message: "User not found in database",
			})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("userID", dbUser.ID)
		c.Set("userType", dbUser.UserType)

		// If no roles specified, any authenticated user is allowed
		if len(roles) == 0 {
			c.Next()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if dbUser.UserType == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			sendError(c, ErrorResponse{
				Status:  http.StatusForbidden,
				Code:    ErrForbidden,
				Message: "You don't have permission to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware ensures the user is an admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.UserTypeAdmin)
}

// RequireDriver middleware ensures the user is a driver
func RequireDriver() gin.HandlerFunc {
	return RequireRole(models.UserTypeDriver)
}

// RequireCustomer middleware ensures the user is a customer
func RequireCustomer() gin.HandlerFunc {
	return RequireRole(models.UserTypeCustomer)
}

// RequireAdminOrDriver middleware allows both admins and drivers
func RequireAdminOrDriver() gin.HandlerFunc {
	return RequireRole(models.UserTypeAdmin, models.UserTypeDriver)
}

// ReadUserTypeFromContext gets the user's type from the context
func ReadUserTypeFromContext(c *gin.Context) (models.UserType, bool) {
	// Get the user ID from context
	userIDValue, exists := c.Get("userID")
	if !exists {
		return "", false
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		return "", false
	}

	// Get user from database
	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		return "", false
	}

	return user.UserType, true
}
