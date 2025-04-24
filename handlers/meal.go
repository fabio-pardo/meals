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

	// Use transaction to ensure data integrity
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		return tx.Create(&newMeal).Error
	})

	if HandleAppError(c, err) {
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

	// Use transaction to ensure data integrity
	err = store.WithTransaction(c, func(tx *gorm.DB) error {
		// First check if meal exists
		var existingMeal models.Meal
		if result := tx.First(&existingMeal, id); result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Meal"}
			}
			return result.Error
		}
		
		// Then update it
		return tx.Updates(&updatedMeal).Error
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusOK, updatedMeal)
}

func DeleteMealHandler(c *gin.Context) {
	id := c.Param("id")
	var meal models.Meal
	var rowsAffected int64

	// Use transaction to ensure data integrity
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// Check for related records that might be affected
		// (Assuming there's a MenuMeal relation, modify as needed)
		var count int64
		if err := tx.Model(&models.MenuMeal{}).Where("meal_id = ?", id).Count(&count).Error; err != nil {
			return err
		}

		// You could implement your own business rules here
		// For example, you might want to prevent deletion if the meal is part of a menu
		// or cascade delete related records

		result := tx.Delete(&meal, id)
		if result.Error != nil {
			return result.Error
		}
		
		rowsAffected = result.RowsAffected
		return nil
	})

	if HandleAppError(c, err) {
		return
	}
	
	if rowsAffected == 0 {
		RespondWithError(c, NotFoundError("Meal"))
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Meal successfully deleted"})
}
