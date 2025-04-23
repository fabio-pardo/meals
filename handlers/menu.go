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
		RespondWithError(c, BadRequestError("Invalid or malformed menu data"))
		return
	}

	if err := store.DB.Create(&newMenu).Error; err != nil {
		RespondWithError(c, DatabaseError("Failed to create menu"))
		return
	}

	c.JSON(http.StatusCreated, newMenu)
}

func UpdateMenuHandler(c *gin.Context) {
	var updatedMenu models.Menu
	if err := c.BindJSON(&updatedMenu); err != nil {
		RespondWithError(c, BadRequestError("Invalid or malformed menu data"))
		return
	}
	if err := store.DB.Updates(&updatedMenu).Error; err != nil {
		RespondWithError(c, DatabaseError("Failed to update menu"))
		return
	}
	c.JSON(http.StatusOK, updatedMenu)
}
