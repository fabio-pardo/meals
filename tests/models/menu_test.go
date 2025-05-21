package models_test

import (
	"meals/models"
	"meals/tests/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMenuCreation(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Arrange
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)

	testMenu := models.Menu{
		Name:          "Test Weekly Menu",
		Description:   "A test menu for this week",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}

	// Act
	result := db.Create(&testMenu)

	// Assert
	assert.Nil(t, result.Error)
	assert.NotZero(t, testMenu.ID)
	assert.NotZero(t, testMenu.CreatedAt)
	assert.NotZero(t, testMenu.UpdatedAt)

	// Verify menu was saved correctly
	var retrievedMenu models.Menu
	db.First(&retrievedMenu, testMenu.ID)

	assert.Equal(t, testMenu.Name, retrievedMenu.Name)
	assert.Equal(t, testMenu.Description, retrievedMenu.Description)
	assert.Equal(t, testMenu.WeekStartDate.Unix(), retrievedMenu.WeekStartDate.Unix())
	assert.Equal(t, testMenu.WeekEndDate.Unix(), retrievedMenu.WeekEndDate.Unix())
}

func TestMenuUpdate(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a menu
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)
	menu := models.Menu{
		Name:          "Original Menu",
		Description:   "Original description",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}
	db.Create(&menu)

	// Update the menu
	newStart := weekStart.AddDate(0, 0, 7)
	newEnd := newStart.AddDate(0, 0, 7)
	menu.Name = "Updated Menu"
	menu.Description = "Updated description"
	menu.WeekStartDate = newStart
	menu.WeekEndDate = newEnd
	db.Save(&menu)

	// Verify updates
	var updatedMenu models.Menu
	db.First(&updatedMenu, menu.ID)

	assert.Equal(t, "Updated Menu", updatedMenu.Name)
	assert.Equal(t, "Updated description", updatedMenu.Description)
	assert.Equal(t, newStart.Unix(), updatedMenu.WeekStartDate.Unix())
	assert.Equal(t, newEnd.Unix(), updatedMenu.WeekEndDate.Unix())
}

func TestMenuDelete(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a menu
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 7)
	menu := models.Menu{
		Name:          "Menu to Delete",
		Description:   "Will be deleted",
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
	}
	db.Create(&menu)

	// Delete the menu
	result := db.Delete(&menu)
	assert.Nil(t, result.Error)

	// Verify menu is deleted (soft delete with GORM)
	var deletedMenu models.Menu
	result = db.First(&deletedMenu, menu.ID)
	assert.Error(t, result.Error) // Should not find the menu

	// Try with Unscoped to see if it's really in the DB (soft deleted)
	result = db.Unscoped().First(&deletedMenu, menu.ID)
	assert.Nil(t, result.Error)             // Should find the menu in unscoped query
	assert.NotNil(t, deletedMenu.DeletedAt) // Should have a deletion timestamp
}

func TestMenuQuery(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create multiple menus for different weeks
	now := time.Now()
	menus := []models.Menu{
		{
			Name:          "Week 1 Menu",
			Description:   "First week",
			WeekStartDate: now,
			WeekEndDate:   now.AddDate(0, 0, 7),
		},
		{
			Name:          "Week 2 Menu",
			Description:   "Second week",
			WeekStartDate: now.AddDate(0, 0, 7),
			WeekEndDate:   now.AddDate(0, 0, 14),
		},
		{
			Name:          "Week 3 Menu",
			Description:   "Third week",
			WeekStartDate: now.AddDate(0, 0, 14),
			WeekEndDate:   now.AddDate(0, 0, 21),
		},
	}
	for _, menu := range menus {
		db.Create(&menu)
	}

	// Test querying by date range
	var futureMenus []models.Menu
	futureDate := now.AddDate(0, 0, 10) // A date in the second week
	result := db.Where("week_start_date <= ? AND week_end_date >= ?", futureDate, futureDate).Find(&futureMenus)
	assert.Nil(t, result.Error)
	assert.Equal(t, 1, len(futureMenus)) // Only Week 2 Menu should be returned

	// Test querying by name
	var week1Menu models.Menu
	result = db.Where("name = ?", "Week 1 Menu").First(&week1Menu)
	assert.Nil(t, result.Error)
	assert.Equal(t, "Week 1 Menu", week1Menu.Name)
}
