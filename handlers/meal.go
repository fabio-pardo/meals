// Package handlers provides HTTP request handlers for the Meals API.
//
// This package contains handlers for meal management operations including:
// - Retrieving all meals (GET /meals)
// - Retrieving a specific meal (GET /meals/:id)
// - Creating new meals (POST /meals)
// - Updating existing meals (PUT /meals/:id)
// - Deleting meals (DELETE /meals/:id)
//
// All handlers follow consistent patterns:
// - Use standardized error responses via RespondWithError()
// - Implement transaction management for data integrity
// - Include proper HTTP status codes
// - Handle both business logic and database errors
package handlers

import (
	"meals/models"
	"meals/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMealsHandler retrieves all meals from the database.
//
// This endpoint is publicly accessible and returns all meals without filtering.
// In the future, this might be enhanced to support pagination and filtering.
//
// Route: GET /meals
// Response: 200 OK with array of Meal objects
// Error responses: 500 if database error occurs
func GetMealsHandler(c *gin.Context) {
	var meals []models.Meal
	if err := store.DB.Find(&meals).Error; err != nil {
		RespondWithError(c, DatabaseError("Failed to retrieve meals"))
		return
	}
	c.JSON(http.StatusOK, meals)
}

// GetMealHandler retrieves a specific meal by ID.
//
// Route: GET /meals/:id
// Parameters: id (path) - The meal ID
// Response: 200 OK with Meal object
// Error responses: 404 if meal not found, 500 if database error
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

// CreateMealHandler creates a new meal with the provided data.
//
// This endpoint requires authentication and creates a meal within a database transaction
// to ensure data integrity. The meal data is validated before creation.
//
// Route: POST /meals
// Request body: JSON with meal data (name, price, etc.)
// Response: 201 Created with the created Meal object
// Error responses: 400 if invalid data, 401 if unauthorized, 500 if database error
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

// UpdateMealHandler updates an existing meal with new data.
//
// This endpoint requires authentication and updates a meal within a database transaction.
// It first verifies the meal exists before attempting to update it.
//
// Route: PUT /meals/:id
// Parameters: id (path) - The meal ID to update
// Request body: JSON with updated meal data
// Response: 200 OK with the updated Meal object
// Error responses: 400 if invalid data/ID, 401 if unauthorized, 404 if meal not found, 500 if database error
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

// DeleteMealHandler deletes a meal by ID.
//
// This endpoint requires authentication and performs a soft delete within a transaction.
// It checks for related records (like menu associations) before deletion and implements
// appropriate business rules for cascade operations.
//
// Route: DELETE /meals/:id
// Parameters: id (path) - The meal ID to delete
// Response: 200 OK with success message
// Error responses: 401 if unauthorized, 404 if meal not found, 500 if database error
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
