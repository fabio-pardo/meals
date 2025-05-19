package tests

import (
	"log"
	"meals/config"
	"meals/models"
	"meals/store"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var TestDB *gorm.DB

// InitTestDB initializes a test database connection
func InitTestDB() *gorm.DB {
	// Use test environment
	os.Setenv("APP_ENV", "test")
	config.InitConfig()

	// Configure test database logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info, // Increased verbosity for debugging
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Connect to test database
	dbConfig := config.AppConfig.Database

	// Override with environment variables if present
	if os.Getenv("DATABASE_USER") != "" {
		dbConfig.User = os.Getenv("DATABASE_USER")
	}
	if os.Getenv("DATABASE_PASSWORD") != "" {
		dbConfig.Password = os.Getenv("DATABASE_PASSWORD")
	}
	if os.Getenv("DATABASE_NAME") != "" {
		dbConfig.Name = os.Getenv("DATABASE_NAME")
	}

	dsn := dbConfig.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true, // Enable foreign key constraints for testing
	})

	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Disable foreign key checks temporarily
	err = db.Exec("SET CONSTRAINTS ALL DEFERRED").Error
	if err != nil {
		log.Fatalf("Failed to defer constraints: %v", err)
	}

	// Drop all tables to ensure clean state
	tables := []string{
		"order_items",
		"orders",
		"addresses",
		"user_profiles",
		"menu_meals",
		"menus",
		"users",
		"meals",
	}

	for _, table := range tables {
		err = db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error
		if err != nil {
			log.Fatalf("Failed to drop table %s: %v", table, err)
		}
	}

	// Disable foreign key checks during migration
	err = db.Exec("SET CONSTRAINTS ALL DEFERRED").Error
	if err != nil {
		log.Fatalf("Failed to disable constraints: %v", err)
	}

	// Migrate tables in dependency order
	err = db.AutoMigrate(
		&models.User{},
		&models.Order{},
		&models.OrderItem{},
		&models.Meal{},
		&models.Menu{},
		&models.MenuMeal{},
		&models.UserProfile{},
		&models.Address{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate tables: %v", err)
	}

	// Re-enable foreign key checks
	err = db.Exec("SET CONSTRAINTS ALL IMMEDIATE").Error
	if err != nil {
		log.Fatalf("Failed to re-enable constraints: %v", err)
	}

	return db
}

// ClearTestDB clears all data from the test database
func ClearTestDB(db *gorm.DB) error {
	// Delete all data in reverse order of dependencies
	tables := []string{
		"order_items",
		"orders",
		"addresses",
		"user_profiles",
		"menu_meals",
		"menus",
		"meals",
		"users",
	}

	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			return err
		}
	}

	return nil
}

// SetupTestSuite sets up the test suite
func SetupTestSuite(t *testing.T) *gorm.DB {
	db := InitTestDB()
	// Set global store DB for transaction tests
	store.DB = db
	// Drop custom test tables to ensure a clean slate
	if err := db.Exec("DROP TABLE IF EXISTS transaction_success_test, transaction_fail_test CASCADE").Error; err != nil {
		t.Fatalf("Failed to drop custom test tables: %v", err)
	}
	err := ClearTestDB(db)
	if err != nil {
		t.Fatalf("Failed to clear test database: %v", err)
	}
	return db
}

// SetupTest sets up each test
func SetupTest(t *testing.T, db *gorm.DB) {
	err := ClearTestDB(db)
	if err != nil {
		t.Fatalf("Failed to clear test database: %v", err)
	}
}

// CreateTestUser creates a test user
func CreateTestUser(db *gorm.DB, userType models.UserType) models.User {
	timestamp := time.Now().Format("20060102150405")
	user := models.User{
		Provider:    "google",
		Email:       "test" + timestamp + "@example.com",
		Name:        "Test User " + timestamp,
		FirstName:   "Test",
		LastName:    "User " + timestamp,
		UserID:      "test_user_" + timestamp,
		AccessToken: "test_token_" + timestamp,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		IDToken:     "test_id_token_" + timestamp,
		UserType:    userType,
	}

	// Create user and check for errors
	result := db.Create(&user)
	if result.Error != nil {
		log.Printf("Error creating test user: %v", result.Error)
	}

	// Ensure the user has a valid ID
	if user.ID == 0 {
		// Try to find the user by email if create didn't set the ID
		db.Where("email = ?", user.Email).First(&user)
	}

	return user
}

// CreateTestMeal creates a test meal
func CreateTestMeal(db *gorm.DB, name string, price float64) models.Meal {
	meal := models.Meal{
		Name:  name,
		Price: price,
	}

	db.Create(&meal)
	return meal
}

// CreateTestMenu creates a test menu with meals
func CreateTestMenu(db *gorm.DB, name string, mealIDs []uint) models.Menu {
	menu := models.Menu{
		Name:          name,
		Description:   "Test menu description",
		WeekStartDate: time.Now(),
		WeekEndDate:   time.Now().Add(7 * 24 * time.Hour),
	}

	db.Create(&menu)

	// Create menu-meal associations
	for _, mealID := range mealIDs {
		menuMeal := models.MenuMeal{
			MenuID:      menu.ID,
			MealID:      mealID,
			DeliveryDay: "Monday",
		}
		db.Create(&menuMeal)
	}

	return menu
}

// CreateTestOrder creates a test order
func CreateTestOrder(db *gorm.DB, userID uint, items []models.OrderItem) models.Order {
	order := models.Order{
		UserID:          userID,
		Status:          models.OrderStatusPending,
		DeliveryAddress: "123 Test St, Test City, Test State 12345",
		DeliveryDate:    time.Now().Add(24 * time.Hour),
		PaymentMethod:   "credit_card",
	}

	db.Create(&order)

	// Add order items with the order ID
	for i := range items {
		items[i].OrderID = order.ID
		db.Create(&items[i])
	}

	// Calculate and update total amount
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = total
	db.Save(&order)

	return order
}

// CreateTestProfile creates a test user profile
func CreateTestProfile(db *gorm.DB, userID uint) models.UserProfile {
	profile := models.UserProfile{
		UserID:      userID,
		PhoneNumber: "555-123-4567",
	}

	db.Create(&profile)
	return profile
}

// CreateTestAddress creates a test address
func CreateTestAddress(db *gorm.DB, profileID uint, isDefault bool) models.Address {
	address := models.Address{
		UserProfileID: profileID,
		Name:          "Home",
		Street:        "123 Test St",
		City:          "Test City",
		State:         "Test State",
		ZipCode:       "12345",
		Country:       "USA",
		IsDefault:     isDefault,
	}

	db.Create(&address)

	// Update profile's default address if this is the default
	if isDefault {
		db.Model(&models.UserProfile{}).Where("id = ?", profileID).Update("default_address_id", address.ID)
	}

	return address
}
