package tests

import (
	"meals/models"
	"testing"
	"time"

	"meals/tests/testutils"

	"github.com/stretchr/testify/assert"
)

// TestConnect ensures the test database is working
func TestConnect(t *testing.T) {
	// Set up once
	DB := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(DB)

	// Create a test user
	user := models.User{
		Provider:    "google",
		Email:       "test@example.com",
		Name:        "Test User",
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(600),
		IDToken:     "test-id-token",
		UserID:      "123456789",
		UserType:    models.UserTypeCustomer,
	}

	// Act - save to database
	result := DB.Create(&user)

	// Assert it worked
	assert.NoError(t, result.Error)
	assert.NotZero(t, user.ID)

	// Try to retrieve it
	var found models.User
	result = DB.First(&found, user.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, user.Email, found.Email)
}
