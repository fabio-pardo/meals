package controllers

import (
	"meals/config"
	"meals/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMeals(c *gin.Context) {
	var meals []models.Meal
	if err := config.DB.Find(&meals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meals"})
		return
	}
	c.JSON(http.StatusOK, meals)
}

func GetMealByID(c *gin.Context) {
	id := c.Param("id")
	var meal models.Meal

	if err := config.DB.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "meal not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meal"})
		}
		return
	}

	c.JSON(http.StatusOK, meal)
}

func PostMeals(c *gin.Context) {
	var newMeal models.Meal

	if err := c.BindJSON(&newMeal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := config.DB.Create(&newMeal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meal"})
		return
	}

	c.JSON(http.StatusCreated, newMeal)
}
