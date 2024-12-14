package handlers

import (
	"meals/models"
	"meals/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMealsHandler(c *gin.Context) {
	var meals []models.Meal
	if err := store.DB.Find(&meals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meals"})
		return
	}
	c.JSON(http.StatusOK, meals)
}

func GetMealHandler(c *gin.Context) {
	id := c.Param("id")
	var meal models.Meal

	if err := store.DB.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "meal not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meal"})
		}
		return
	}

	c.JSON(http.StatusOK, meal)
}

func CreateMealHandler(c *gin.Context) {
	var newMeal models.Meal

	if err := c.BindJSON(&newMeal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := store.DB.Create(&newMeal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meal"})
		return
	}

	c.JSON(http.StatusCreated, newMeal)
}

func UpdateMealHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updatedMeal models.Meal
	updatedMeal.ID = uint(id)

	if err := c.BindJSON(&updatedMeal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := store.DB.Updates(&updatedMeal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store updated meal"})
		return
	}

	c.JSON(http.StatusCreated, updatedMeal)
}

func DeleteMealHandler(c *gin.Context) {
	id := c.Param("id")
	var meal models.Meal

	deletedMeal := store.DB.Delete(&meal, id)
	if deletedMeal.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete meal"})
		return
	}
	if deletedMeal.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Could not find meal for deletion"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Meal successfully deleted"})
}
