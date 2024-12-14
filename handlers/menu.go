package handlers

import (
	"meals/models"
	"meals/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateMenuHandler(c *gin.Context) {
	var newMenu models.Menu
	if err := c.BindJSON(&newMenu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := store.DB.Create(&newMenu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
		return
	}

	c.JSON(http.StatusCreated, newMenu)
}

func UpdateMenuHandler(c *gin.Context) {
	var updatedMenu models.Menu
	if err := c.BindJSON(&updatedMenu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	if err := store.DB.Updates(&updatedMenu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
		return
	}
	c.JSON(http.StatusCreated, updatedMenu)
}

func GetMenuHandler(c *gin.Context) {
	var menus []models.Menu
	if err := store.DB.Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menus"})
		return
	}

	c.JSON(http.StatusOK, menus)
}
