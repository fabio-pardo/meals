package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

func HomeHandler(c *gin.Context) {
	gUser, _ := c.Get("user")
	if gUser, ok := gUser.(goth.User); ok {
		c.String(http.StatusOK, fmt.Sprintf("Welcome %v", gUser.Email))
	} else {
		c.String(http.StatusOK, "Welcome! Go to /auth/google to log in with Google.")
	}
}
