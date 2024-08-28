package controllers

import (
	"meals/config"
	"meals/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PostMenu(c *gin.Context) {
	var newMenu models.Menu
	if err := c.BindJSON(&newMenu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := config.DB.Create(&newMenu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
		return
	}

	c.JSON(http.StatusCreated, newMenu)
}
