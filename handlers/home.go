package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

func HomeHandler(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		// User not found in context, but this is still an OK response for home page
		c.String(http.StatusOK, "Welcome! Go to /auth/google to log in with Google.")
		return
	}

	if gothUser, ok := userValue.(goth.User); ok {
		c.String(http.StatusOK, fmt.Sprintf("Welcome %v", gothUser.Email))
	} else {
		// User exists in context but has wrong type - this should never happen
		log.Printf("Expected goth.User in context, got %T", userValue)
		c.String(http.StatusOK, "Welcome! Go to /auth/google to log in with Google.")
	}
}
