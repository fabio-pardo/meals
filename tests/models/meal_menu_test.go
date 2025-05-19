package models_test

import (
	"meals/models"
	"meals/tests"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMealModel(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("CreateMeal", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a new meal
		meal := models.Meal{
			Name:  "Test Chicken Bowl",
			Price: 12.99,
		}

		// Save the meal
		err := db.Create(&meal).Error
		assert.Nil(t, err, "Expected no error when creating meal")
		assert.NotZero(t, meal.ID, "Expected meal ID to be assigned")

		// Retrieve the meal
		var retrievedMeal models.Meal
		err = db.First(&retrievedMeal, meal.ID).Error
		assert.Nil(t, err, "Expected no error when retrieving meal")
		assert.Equal(t, meal.Name, retrievedMeal.Name)
		assert.Equal(t, meal.Price, retrievedMeal.Price)
	})

	t.Run("UpdateMeal", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a meal using helper
		meal := tests.CreateTestMeal(db, "Initial Meal", 10.99)

		// Update the meal
		meal.Name = "Updated Meal"
		meal.Price = 14.99

		err := db.Save(&meal).Error
		assert.Nil(t, err, "Expected no error when updating meal")

		// Retrieve the updated meal
		var retrievedMeal models.Meal
		err = db.First(&retrievedMeal, meal.ID).Error
		assert.Nil(t, err, "Expected no error when retrieving updated meal")
		assert.Equal(t, "Updated Meal", retrievedMeal.Name)
		assert.Equal(t, 14.99, retrievedMeal.Price)

	})

	t.Run("DeleteMeal", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a meal
		meal := tests.CreateTestMeal(db, "Meal to Delete", 11.99)

		// Delete the meal
		err := db.Delete(&meal).Error
		assert.Nil(t, err, "Expected no error when deleting meal")

		// Try to retrieve the deleted meal
		var retrievedMeal models.Meal
		err = db.First(&retrievedMeal, meal.ID).Error
		assert.Error(t, err, "Expected error when retrieving deleted meal")
	})
}

func TestMenuModel(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("CreateMenu", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create meals for the menu
		meal1 := tests.CreateTestMeal(db, "Meal 1", 9.99)
		meal2 := tests.CreateTestMeal(db, "Meal 2", 11.99)
		meal3 := tests.CreateTestMeal(db, "Meal 3", 13.99)

		// Create a menu
		startDate := time.Now()
		endDate := startDate.Add(7 * 24 * time.Hour)
		menu := models.Menu{
			Name:          "Weekly Special Menu",
			Description:   "Our special selections for this week",
			WeekStartDate: startDate,
			WeekEndDate:   endDate,
		}

		// Save the menu
		err := db.Create(&menu).Error
		assert.Nil(t, err, "Expected no error when creating menu")

		// Create menu-meal associations
		for _, mealID := range []uint{meal1.ID, meal2.ID, meal3.ID} {
			menuMeal := models.MenuMeal{
				MenuID:      menu.ID,
				MealID:      mealID,
				DeliveryDay: "Monday",
			}
			db.Create(&menuMeal)
		}

		// Retrieve the menu with menu meals
		var retrievedMenu models.Menu
		err = db.Preload("MenuMeals").First(&retrievedMenu, menu.ID).Error
		assert.Nil(t, err, "Expected no error when retrieving menu with meals")
		assert.Equal(t, menu.Name, retrievedMenu.Name)
		assert.Equal(t, menu.Description, retrievedMenu.Description)
		assert.Len(t, retrievedMenu.MenuMeals, 3, "Expected menu to have 3 menu meals")
	})

	t.Run("UpdateMenu", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create meals
		meal1 := tests.CreateTestMeal(db, "Meal 1", 9.99)
		meal2 := tests.CreateTestMeal(db, "Meal 2", 11.99)

		// Create a menu with initial meals
		menu := tests.CreateTestMenu(db, "Initial Menu", []uint{meal1.ID})

		// Update the menu
		menu.Name = "Updated Menu"
		menu.Description = "Updated description"

		err := db.Save(&menu).Error
		assert.Nil(t, err, "Expected no error when updating menu")

		// Create menu-meal association for the new meal
		menuMeal := models.MenuMeal{
			MenuID:      menu.ID,
			MealID:      meal2.ID,
			DeliveryDay: "Tuesday",
		}
		db.Create(&menuMeal)

		// Retrieve the updated menu with menu meals
		var retrievedMenu models.Menu
		err = db.Preload("MenuMeals").First(&retrievedMenu, menu.ID).Error
		assert.Nil(t, err, "Expected no error when retrieving updated menu")
		assert.Equal(t, "Updated Menu", retrievedMenu.Name)
		assert.Equal(t, "Updated description", retrievedMenu.Description)
		assert.Len(t, retrievedMenu.MenuMeals, 2, "Expected menu to have 2 meals after update")
	})

	t.Run("DeleteMenu", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create meals
		meal1 := tests.CreateTestMeal(db, "Meal 1", 9.99)
		meal2 := tests.CreateTestMeal(db, "Meal 2", 11.99)

		// Create a menu
		menu := tests.CreateTestMenu(db, "Menu to Delete", []uint{meal1.ID, meal2.ID})

		// Delete the menu
		err := db.Delete(&menu).Error
		assert.Nil(t, err, "Expected no error when deleting menu")

		// Try to retrieve the deleted menu
		var retrievedMenu models.Menu
		err = db.First(&retrievedMenu, menu.ID).Error
		assert.Error(t, err, "Expected error when retrieving deleted menu")

		// Verify menu-meal associations are deleted (cascade delete)
		var count int64
		db.Model(&models.MenuMeal{}).Where("menu_id = ?", menu.ID).Count(&count)
		assert.Equal(t, int64(0), count, "Expected no menu-meal associations after menu deletion")

		// Verify meals still exist
		var meal1Count, meal2Count int64
		db.Model(&models.Meal{}).Where("id = ?", meal1.ID).Count(&meal1Count)
		db.Model(&models.Meal{}).Where("id = ?", meal2.ID).Count(&meal2Count)
		assert.Equal(t, int64(1), meal1Count, "Expected meal 1 to still exist")
		assert.Equal(t, int64(1), meal2Count, "Expected meal 2 to still exist")
	})
}
