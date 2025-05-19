package handlers

import (
	"meals/models"
	"meals/store"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateMenuHandler(c *gin.Context) {
	var newMenu models.Menu
	if err := c.BindJSON(&newMenu); err != nil {
		RespondWithError(c, BadRequestError("Invalid or malformed menu data"))
		return
	}

	// Use transaction to ensure data integrity
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// Create the menu
		if err := tx.Create(&newMenu).Error; err != nil {
			return err
		}

		// If menu has associated meals, validate and handle the relationships
		if len(newMenu.MenuMeals) > 0 {
			// Verify all referenced meal IDs exist
			var count int64
			if err := tx.Model(&models.Meal{}).Where("id IN ?", newMenu.MenuMeals).Count(&count).Error; err != nil {
				return err
			}

			if int(count) != len(newMenu.MenuMeals) {
				return RelationshipErrorType{
					Message: "One or more meal IDs do not exist",
					Details: map[string]interface{}{
						"provided_ids": newMenu.MenuMeals,
						"found_count":  count,
					},
				}
			}

			// Create menu-meal associations
			for _, meal := range newMenu.MenuMeals {
				menuMeal := models.MenuMeal{
					MenuID: newMenu.ID,
					MealID: meal.MealID,
				}
				if err := tx.Create(&menuMeal).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
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

	if updatedMenu.ID == 0 {
		RespondWithError(c, BadRequestError("Menu ID is required"))
		return
	}

	// Use transaction to ensure data integrity
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// First check if menu exists
		var existingMenu models.Menu
		if result := tx.First(&existingMenu, updatedMenu.ID); result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Menu"}
			}
			return result.Error
		}

		// Update the menu basic properties
		if err := tx.Model(&updatedMenu).Updates(map[string]interface{}{
			"name":        updatedMenu.Name,
			"description": updatedMenu.Description,
			// Add other fields as needed
		}).Error; err != nil {
			return err
		}

		// If meal associations have changed, update them
		if len(updatedMenu.MenuMeals) > 0 {
			// Verify all referenced meal IDs exist
			var count int64
			if err := tx.Model(&models.Meal{}).Where("id IN ?", updatedMenu.MenuMeals).Count(&count).Error; err != nil {
				return err
			}

			if int(count) != len(updatedMenu.MenuMeals) {
				return RelationshipErrorType{
					Message: "One or more meal IDs do not exist",
					Details: map[string]interface{}{
						"provided_ids": updatedMenu.MenuMeals,
						"found_count":  count,
					},
				}
			}

			// Delete existing associations
			if err := tx.Where("menu_id = ?", updatedMenu.ID).Delete(&models.MenuMeal{}).Error; err != nil {
				return err
			}

			// Create new associations
			for _, mealID := range updatedMenu.MenuMeals {
				menuMeal := models.MenuMeal{
					MenuID: updatedMenu.ID,
					MealID: mealID.MealID,
				}
				if err := tx.Create(&menuMeal).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	// Reload the menu to get the updated version
	var refreshedMenu models.Menu
	if err := store.DB.First(&refreshedMenu, updatedMenu.ID).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{Message: "Failed to retrieve updated menu"})
		return
	}

	c.JSON(http.StatusOK, refreshedMenu)
}

// GetMenusHandler retrieves all menus with their associated meals
func GetMenusHandler(c *gin.Context) {
	var menus []models.Menu

	// Use transaction to ensure data consistency
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// Get all menus with their menu-meal associations and the associated meals
		if err := tx.Preload("MenuMeals.Meal").Find(&menus).Error; err != nil {
			return err
		}
		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusOK, menus)
}
