package models_test

import (
	"meals/models"
	"meals/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Define a global test time to use across tests
var testTime = time.Now()

func TestMealCreation(t *testing.T) {

	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Arrange
	testMeal := models.Meal{
		Name:  "Test Meal",
		Price: 12.99,
	}

	// Act
	result := db.Create(&testMeal)

	// Assert
	assert.Nil(t, result.Error)
	assert.NotZero(t, testMeal.ID)
	assert.NotZero(t, testMeal.CreatedAt)

	// Verify meal was saved correctly
	var retrievedMeal models.Meal
	db.First(&retrievedMeal, testMeal.ID)

	assert.Equal(t, testMeal.Name, retrievedMeal.Name)
	assert.Equal(t, testMeal.Price, retrievedMeal.Price)
}

func TestMealUpdate(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a meal
	meal := models.Meal{
		Name:  "Original Meal",
		Price: 9.99,
	}
	db.Create(&meal)

	// Update the meal
	meal.Name = "Updated Meal"
	meal.Price = 14.99
	db.Save(&meal)

	// Verify updates
	var updatedMeal models.Meal
	db.First(&updatedMeal, meal.ID)

	assert.Equal(t, "Updated Meal", updatedMeal.Name)
	assert.Equal(t, 14.99, updatedMeal.Price)
}

func TestMealDelete(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a meal
	meal := models.Meal{
		Name:  "Meal to Delete",
		Price: 19.99,
	}
	db.Create(&meal)

	// Delete the meal
	result := db.Delete(&meal)
	assert.Nil(t, result.Error)

	// Verify meal is deleted (soft delete with GORM)
	var deletedMeal models.Meal
	result = db.First(&deletedMeal, meal.ID)
	assert.Error(t, result.Error) // Should not find the meal

	// Try with Unscoped to see if it's really in the DB (soft deleted)
	result = db.Unscoped().First(&deletedMeal, meal.ID)
	assert.Nil(t, result.Error)             // Should find the meal in unscoped query
	assert.NotNil(t, deletedMeal.DeletedAt) // Should have a deletion timestamp
}

func TestMealQuery(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create multiple meals
	meals := []models.Meal{
		{Name: "Meal 1", Price: 9.99},
		{Name: "Meal 2", Price: 14.99},
		{Name: "Meal 3", Price: 19.99},
	}
	for _, meal := range meals {
		db.Create(&meal)
	}

	// Test querying by price range
	var expensiveMeals []models.Meal
	result := db.Where("price > ?", 10.0).Find(&expensiveMeals)
	assert.Nil(t, result.Error)
	assert.Equal(t, 2, len(expensiveMeals)) // Only Meal 2 and 3 should be returned

	// Test querying by name
	var meal1 models.Meal
	result = db.Where("name = ?", "Meal 1").First(&meal1)
	assert.Nil(t, result.Error)
	assert.Equal(t, "Meal 1", meal1.Name)
	assert.Equal(t, 9.99, meal1.Price)
}
