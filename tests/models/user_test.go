package models_test

import (
	"meals/models"
	"meals/tests/testutils"
	"testing"
	"time"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	// Setup test database
	db = testutils.SetupTestDB()

	// Run tests
	m.Run()

	// Cleanup after tests
	testutils.CleanupTestDB(db)
}

func TestUserCreation(t *testing.T) {
	// Arrange
	testUser := models.User{
		Provider:          "google",
		Email:             "test@example.com",
		Name:              "Test User",
		FirstName:         "Test",
		LastName:          "User",
		AccessToken:       "test-token",
		AccessTokenSecret: "test-secret",
		RefreshToken:      "test-refresh",
		ExpiresAt:         time.Now().Add(time.Hour),
		IDToken:           "test-id-token",
		UserID:            "123456789",
		UserType:          models.UserTypeCustomer,
	}

	// Act
	result := db.Create(&testUser)

	// Assert
	assert.Nil(t, result.Error)
	assert.NotZero(t, testUser.ID)

	// Verify user was saved correctly
	var retrievedUser models.User
	db.First(&retrievedUser, testUser.ID)

	assert.Equal(t, testUser.Email, retrievedUser.Email)
	assert.Equal(t, testUser.Name, retrievedUser.Name)
	assert.Equal(t, testUser.UserID, retrievedUser.UserID)
	assert.Equal(t, models.UserTypeCustomer, retrievedUser.UserType)
}

func TestUserTypeEnums(t *testing.T) {
	// Create users with different types
	adminUser := models.User{
		Provider:    "google",
		Email:       "admin@example.com",
		Name:        "Admin User",
		AccessToken: "admin-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		IDToken:     "admin-id-token",
		UserID:      "admin123",
		UserType:    models.UserTypeAdmin,
	}

	driverUser := models.User{
		Provider:    "google",
		Email:       "driver@example.com",
		Name:        "Driver User",
		AccessToken: "driver-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		IDToken:     "driver-id-token",
		UserID:      "driver123",
		UserType:    models.UserTypeDriver,
	}

	customerUser := models.User{
		Provider:    "google",
		Email:       "customer@example.com",
		Name:        "Customer User",
		AccessToken: "customer-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		IDToken:     "customer-id-token",
		UserID:      "customer123",
		UserType:    models.UserTypeCustomer,
	}

	// Save users
	db.Create(&adminUser)
	db.Create(&driverUser)
	db.Create(&customerUser)

	// Test IsAdmin method
	assert.True(t, adminUser.IsAdmin())
	assert.False(t, driverUser.IsAdmin())
	assert.False(t, customerUser.IsAdmin())

	// Test IsDriver method
	assert.False(t, adminUser.IsDriver())
	assert.True(t, driverUser.IsDriver())
	assert.False(t, customerUser.IsDriver())

	// Test IsCustomer method
	assert.False(t, adminUser.IsCustomer())
	assert.False(t, driverUser.IsCustomer())
	assert.True(t, customerUser.IsCustomer())
}

func TestUserValidation(t *testing.T) {
	// Invalid user with missing required fields
	invalidUser := models.User{
		Name: "Invalid User",
	}

	errors := invalidUser.ValidateUser()
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Email is required")
	assert.Contains(t, errors, "UserID is required")
	assert.Contains(t, errors, "Provider is required")
}

func TestConvertGothUserToModelUser(t *testing.T) {
	// Create a Goth user
	gothUser := &goth.User{
		Provider:          "google",
		Email:             "test@example.com",
		Name:              "Test User",
		FirstName:         "Test",
		LastName:          "User",
		NickName:          "Tester",
		Description:       "A test user",
		UserID:            "google123",
		AvatarURL:         "https://example.com/avatar.jpg",
		AccessToken:       "access-token",
		AccessTokenSecret: "access-token-secret",
		RefreshToken:      "refresh-token",
		ExpiresAt:         time.Now().Add(time.Hour),
		IDToken:           "id-token",
	}

	// Convert to model user
	modelUser, err := models.ConvertGothUserToModelUser(gothUser)

	// Assert conversion succeeded
	assert.NoError(t, err)
	assert.NotNil(t, modelUser)

	// Verify fields were mapped correctly
	assert.Equal(t, gothUser.Provider, modelUser.Provider)
	assert.Equal(t, gothUser.Email, modelUser.Email)
	assert.Equal(t, gothUser.Name, modelUser.Name)
	assert.Equal(t, gothUser.FirstName, modelUser.FirstName)
	assert.Equal(t, gothUser.LastName, modelUser.LastName)
	assert.Equal(t, gothUser.UserID, modelUser.UserID)
	assert.Equal(t, models.UserTypeCustomer, modelUser.UserType) // Default user type should be customer
}
