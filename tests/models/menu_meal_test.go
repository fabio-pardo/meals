package models_test

import (
	"meals/models"
	"meals/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMenuMealCreation(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a meal first
	testMeal := models.Meal{
		Name:  "Menu-Meal Test Item",
		Price: 15.99,
	}
	db.Create(&testMeal)

	// Create a menu
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)
	testMenu := models.Menu{
		Name:          "Menu-Meal Test Menu",
		Description:   "For testing menu-meal relationship",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}
	db.Create(&testMenu)

	// Create the menu-meal association
	testMenuMeal := models.MenuMeal{
		MenuID:      testMenu.ID,
		MealID:      testMeal.ID,
		DeliveryDay: "Monday",
	}

	// Act
	result := db.Create(&testMenuMeal)

	// Assert
	assert.Nil(t, result.Error)
	assert.NotZero(t, testMenuMeal.ID)
	assert.NotZero(t, testMenuMeal.CreatedAt)
	assert.NotZero(t, testMenuMeal.UpdatedAt)

	// Verify menu-meal was saved correctly
	var retrievedMenuMeal models.MenuMeal
	db.First(&retrievedMenuMeal, testMenuMeal.ID)

	assert.Equal(t, testMenu.ID, retrievedMenuMeal.MenuID)
	assert.Equal(t, testMeal.ID, retrievedMenuMeal.MealID)
	assert.Equal(t, "Monday", retrievedMenuMeal.DeliveryDay)
}

func TestMenuMealRelations(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create multiple meals
	meals := []models.Meal{
		{Name: "Breakfast", Price: 8.99},
		{Name: "Lunch", Price: 12.99},
		{Name: "Dinner", Price: 15.99},
	}
	for i := range meals {
		db.Create(&meals[i])
	}

	// Create a menu
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)
	menu := models.Menu{
		Name:          "Full Week Menu",
		Description:   "Menu with multiple meals",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}
	db.Create(&menu)

	// Add meals to the menu for different days
	menuMeals := []models.MenuMeal{
		{MenuID: menu.ID, MealID: meals[0].ID, DeliveryDay: "Monday"},
		{MenuID: menu.ID, MealID: meals[1].ID, DeliveryDay: "Monday"},
		{MenuID: menu.ID, MealID: meals[2].ID, DeliveryDay: "Monday"},
		{MenuID: menu.ID, MealID: meals[0].ID, DeliveryDay: "Tuesday"},
		{MenuID: menu.ID, MealID: meals[1].ID, DeliveryDay: "Tuesday"},
	}
	for _, mm := range menuMeals {
		db.Create(&mm)
	}

	// Test preloading MenuMeals from Menu
	var menuWithMeals models.Menu
	result := db.Preload("MenuMeals").First(&menuWithMeals, menu.ID)
	assert.Nil(t, result.Error)
	assert.Equal(t, 5, len(menuWithMeals.MenuMeals))

	// Test preloading nested relationships (MenuMeals -> Meal)
	var menuWithDetails models.Menu
	result = db.Preload("MenuMeals.Meal").First(&menuWithDetails, menu.ID)
	assert.Nil(t, result.Error)
	assert.Equal(t, 5, len(menuWithDetails.MenuMeals))

	// Verify meal details are loaded
	for _, mm := range menuWithDetails.MenuMeals {
		assert.NotZero(t, mm.Meal.ID)
		assert.NotEmpty(t, mm.Meal.Name)
		assert.NotZero(t, mm.Meal.Price)
	}

	// Test filtering by delivery day
	var mondayMenuMeals []models.MenuMeal
	result = db.Where("menu_id = ? AND delivery_day = ?", menu.ID, "Monday").Find(&mondayMenuMeals)
	assert.Nil(t, result.Error)
	assert.Equal(t, 3, len(mondayMenuMeals))
}

func TestMenuMealCascadeDelete(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a meal
	meal := models.Meal{
		Name:  "Cascade Test Meal",
		Price: 9.99,
	}
	db.Create(&meal)

	// Create a menu
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)
	menu := models.Menu{
		Name:          "Cascade Test Menu",
		Description:   "For testing cascade delete",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}
	db.Create(&menu)

	// Create menu-meal association
	menuMeal := models.MenuMeal{
		MenuID:      menu.ID,
		MealID:      meal.ID,
		DeliveryDay: "Friday",
	}
	db.Create(&menuMeal)

	// Verify menu-meal was created
	var check models.MenuMeal
	result := db.First(&check, menuMeal.ID)
	assert.Nil(t, result.Error)

	// Delete the menu (should cascade to menu-meals)
	db.Delete(&menu)

	// Check if menu-meal was also deleted
	var afterDelete models.MenuMeal
	result = db.First(&afterDelete, menuMeal.ID)
	assert.Error(t, result.Error) // Should not find it

	// Verify meal still exists (no cascade to meal)
	var mealCheck models.Meal
	result = db.First(&mealCheck, meal.ID)
	assert.Nil(t, result.Error) // Should still find the meal
}

func TestMenuWithMultipleMeals(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a week menu with three days of meals
	// This tests creating a complete menu structure

	// 1. Create meals
	meals := []models.Meal{
		{Name: "Oatmeal", Price: 5.99},
		{Name: "Sandwich", Price: 7.99},
		{Name: "Salad", Price: 8.99},
		{Name: "Pasta", Price: 10.99},
		{Name: "Steak", Price: 18.99},
		{Name: "Soup", Price: 6.99},
	}
	for i := range meals {
		db.Create(&meals[i])
	}

	// 2. Create a menu
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)
	menu := models.Menu{
		Name:          "Three-Day Menu",
		Description:   "Complete test menu",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}
	db.Create(&menu)

	// 3. Create menu-meal associations
	// Monday: breakfast, lunch, dinner
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[0].ID, DeliveryDay: "Monday"})
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[1].ID, DeliveryDay: "Monday"})
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[4].ID, DeliveryDay: "Monday"})

	// Tuesday: breakfast, lunch, dinner
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[0].ID, DeliveryDay: "Tuesday"})
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[2].ID, DeliveryDay: "Tuesday"})
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[3].ID, DeliveryDay: "Tuesday"})

	// Wednesday: breakfast, lunch, dinner
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[0].ID, DeliveryDay: "Wednesday"})
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[5].ID, DeliveryDay: "Wednesday"})
	db.Create(&models.MenuMeal{MenuID: menu.ID, MealID: meals[4].ID, DeliveryDay: "Wednesday"})

	// 4. Test retrieving the complete menu with all meals
	var completeMenu models.Menu
	result := db.Preload("MenuMeals.Meal").First(&completeMenu, menu.ID)
	assert.Nil(t, result.Error)
	assert.Equal(t, 9, len(completeMenu.MenuMeals))

	// 5. Test grouping by day
	days := map[string]int{
		"Monday":    0,
		"Tuesday":   0,
		"Wednesday": 0,
	}

	for _, mm := range completeMenu.MenuMeals {
		days[mm.DeliveryDay]++
	}

	assert.Equal(t, 3, days["Monday"])
	assert.Equal(t, 3, days["Tuesday"])
	assert.Equal(t, 3, days["Wednesday"])
}
