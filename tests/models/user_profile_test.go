package models_test

import (
	"meals/models"
	"meals/tests/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserProfileCreation(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a test user first
	testUser := models.User{
		Provider:    "google",
		Email:       "profile-test@example.com",
		Name:        "Profile Test",
		AccessToken: "profile-token",
		ExpiresAt:   testTime,
		IDToken:     "profile-id-token",
		UserID:      "profile123",
		UserType:    models.UserTypeCustomer,
	}
	db.Create(&testUser)

	// Create a user profile
	testProfile := models.UserProfile{
		UserID: testUser.ID,
	}

	// Save the profile
	result := db.Create(&testProfile)
	assert.Nil(t, result.Error)
	assert.NotZero(t, testProfile.ID)

	// Verify profile was saved correctly
	var retrievedProfile models.UserProfile
	db.First(&retrievedProfile, testProfile.ID)
	assert.Equal(t, testUser.ID, retrievedProfile.UserID)

	// Test loading user relation with preload
	var profileWithUser models.UserProfile
	db.Preload("User").First(&profileWithUser, testProfile.ID)
	assert.Equal(t, testUser.ID, profileWithUser.User.ID)
	assert.Equal(t, testUser.Email, profileWithUser.User.Email)
}

func TestUserProfileAssociation(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a test user
	testUser := models.User{
		Provider:    "google",
		Email:       "assoc-test@example.com",
		Name:        "Association Test",
		AccessToken: "assoc-token",
		ExpiresAt:   testTime,
		IDToken:     "assoc-id-token",
		UserID:      "assoc123",
		UserType:    models.UserTypeCustomer,
	}
	db.Create(&testUser)

	// Create a user profile
	testProfile := models.UserProfile{
		UserID: testUser.ID,
	}
	db.Create(&testProfile)

	// Test finding profile by user ID
	var userProfile models.UserProfile
	result := db.Where("user_id = ?", testUser.ID).First(&userProfile)
	assert.Nil(t, result.Error)
	assert.Equal(t, testProfile.ID, userProfile.ID)

	// Update user and check that foreign key works
	testUser.Name = "Updated Name"
	db.Save(&testUser)

	// Reload profile with user
	db.Preload("User").First(&userProfile, userProfile.ID)
	assert.Equal(t, "Updated Name", userProfile.User.Name)
}

// Test deleting a user and ensuring cascading behavior works as expected
func TestUserProfileCascade(t *testing.T) {
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	// Create a test user
	cascadeUser := models.User{
		Provider:    "google",
		Email:       "cascade@example.com",
		Name:        "Cascade Test",
		AccessToken: "cascade-token",
		ExpiresAt:   testTime,
		IDToken:     "cascade-id-token",
		UserID:      "cascade123",
		UserType:    models.UserTypeCustomer,
	}
	db.Create(&cascadeUser)

	// Create a user profile
	cascadeProfile := models.UserProfile{
		UserID: cascadeUser.ID,
	}
	db.Create(&cascadeProfile)

	// Verify profile exists
	var profileCheck models.UserProfile
	resultBefore := db.First(&profileCheck, cascadeProfile.ID)
	assert.Nil(t, resultBefore.Error)

	// Delete the user
	db.Delete(&cascadeUser)

	// Check that the profile still exists (since we don't have ON DELETE CASCADE)
	// The User-UserProfile relationship is not configured with cascade delete
	var profileAfter models.UserProfile
	resultAfter := db.First(&profileAfter, cascadeProfile.ID)
	assert.Nil(t, resultAfter.Error) // Profile should still exist
}
