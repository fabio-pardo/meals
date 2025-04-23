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
		RespondWithError(c, DatabaseError("Failed to retrieve meals"))
		return
	}
	c.JSON(http.StatusOK, meals)
}

func GetMealHandler(c *gin.Context) {
	id := c.Param("id")
	var meal models.Meal

	if err := store.DB.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(c, NotFoundError("Meal"))
		} else {
			RespondWithError(c, DatabaseError("Failed to retrieve meal"))
		}
		return
	}

	c.JSON(http.StatusOK, meal)
}

func CreateMealHandler(c *gin.Context) {
	var newMeal models.Meal

	if err := c.BindJSON(&newMeal); err != nil {
		RespondWithError(c, BadRequestError("Invalid or malformed meal data"))
		return
	}

	if err := store.DB.Create(&newMeal).Error; err != nil {
		RespondWithError(c, DatabaseError("Failed to create meal"))
		return
	}

	c.JSON(http.StatusCreated, newMeal)
}

func UpdateMealHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid meal ID format"))
		return
	}
	
	var updatedMeal models.Meal
	updatedMeal.ID = uint(id)

	if err := c.BindJSON(&updatedMeal); err != nil {
		RespondWithError(c, BadRequestError("Invalid or malformed meal data"))
		return
	}

	if err := store.DB.Updates(&updatedMeal).Error; err != nil {
		RespondWithError(c, DatabaseError("Failed to update meal"))
		return
	}

	c.JSON(http.StatusOK, updatedMeal)
}

func DeleteMealHandler(c *gin.Context) {
	id := c.Param("id")
	var meal models.Meal

	deletedMeal := store.DB.Delete(&meal, id)
	if deletedMeal.Error != nil {
		RespondWithError(c, DatabaseError("Failed to delete meal"))
		return
	}
	if deletedMeal.RowsAffected == 0 {
		RespondWithError(c, NotFoundError("Meal"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Meal successfully deleted"})
}
