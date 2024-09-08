package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomeHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome! Go to /auth/google to log in with Google.")
}
