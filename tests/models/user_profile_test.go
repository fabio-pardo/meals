package models_test

import (
	"fmt"
	"meals/models"
	"meals/tests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserProfileManagement(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("CreateUserProfile", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a profile for the user
		profile := models.UserProfile{
			UserID:             fmt.Sprintf("%d", user.ID),
			PhoneNumber:        "555-123-4567",
			DietaryPreferences: []string{"vegetarian", "nuts-free"},
			PreferencesJSON:    `{"vegetarian": true, "allergies": ["nuts", "dairy"]}`,
		}

		// Save the profile
		err := db.Create(&profile).Error
		assert.Nil(t, err, "Expected no error when creating profile")

		// Retrieve the profile
		var retrievedProfile models.UserProfile
		err = db.Where("user_id = ?", user.ID).First(&retrievedProfile).Error
		assert.Nil(t, err, "Expected no error when retrieving profile")
		assert.Equal(t, profile.PhoneNumber, retrievedProfile.PhoneNumber)
		assert.Equal(t, profile.DietaryPreferences, retrievedProfile.DietaryPreferences)
	})

	t.Run("UpdateUserProfile", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a profile using the helper function
		profile := tests.CreateTestProfile(db, user.ID)

		// Update the profile
		profile.PhoneNumber = "555-987-6543"
		profile.DietaryPreferences = []string{"vegan", "gluten-free"}
		profile.PreferencesJSON = `{"vegan": true, "allergies": ["gluten"]}`

		err := db.Save(&profile).Error
		assert.Nil(t, err, "Expected no error when updating profile")

		// Retrieve the updated profile
		var retrievedProfile models.UserProfile
		err = db.Where("user_id = ?", user.ID).First(&retrievedProfile).Error
		assert.Nil(t, err, "Expected no error when retrieving updated profile")
		assert.Equal(t, "555-987-6543", retrievedProfile.PhoneNumber)
		assert.Equal(t, `{"vegan": true, "allergies": ["gluten"]}`, retrievedProfile.PreferencesJSON)
		assert.ElementsMatch(t, []string{"vegan", "gluten-free"}, retrievedProfile.DietaryPreferences)
	})

	t.Run("DeleteUserProfile", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a profile
		profile := tests.CreateTestProfile(db, user.ID)

		// Delete the profile
		err := db.Delete(&profile).Error
		assert.Nil(t, err, "Expected no error when deleting profile")

		// Try to retrieve the deleted profile
		var retrievedProfile models.UserProfile
		err = db.Where("user_id = ?", user.ID).First(&retrievedProfile).Error
		assert.Error(t, err, "Expected error when retrieving deleted profile")
	})
}

func TestAddressManagement(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("CreateAddress", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user and profile
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		profile := tests.CreateTestProfile(db, user.ID)

		// Create an address for the profile
		address := models.Address{
			UserProfileID: fmt.Sprintf("%d", profile.ID),
			Name:          "Home",
			Street:        "123 Test St",
			City:          "Test City",
			State:         "Test State",
			ZipCode:       "12345",
			Country:       "USA",
			IsDefault:     true,
		}

		// Save the address
		err := db.Create(&address).Error
		assert.Nil(t, err, "Expected no error when creating address")

		// Verify the address was created
		var retrievedAddress models.Address
		err = db.Where("user_profile_id = ?", fmt.Sprintf("%d", profile.ID)).First(&retrievedAddress).Error
		assert.Nil(t, err, "Expected no error when retrieving address")
		assert.Equal(t, address.Street, retrievedAddress.Street)
		assert.Equal(t, address.City, retrievedAddress.City)
		assert.True(t, retrievedAddress.IsDefault, "Expected address to be default")
	})

	t.Run("UpdateAddress", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user and profile
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		profile := tests.CreateTestProfile(db, user.ID)

		// Create an address using helper function
		address := tests.CreateTestAddress(db, profile.ID, true)

		// Update the address
		address.Street = "456 New St"
		address.City = "New City"

		err := db.Save(&address).Error
		assert.Nil(t, err, "Expected no error when updating address")

		// Verify the address was updated
		var retrievedAddress models.Address
		err = db.Where("id = ?", address.ID).First(&retrievedAddress).Error
		assert.Nil(t, err, "Expected no error when retrieving updated address")
		assert.Equal(t, "456 New St", retrievedAddress.Street)
		assert.Equal(t, "New City", retrievedAddress.City)
	})

	t.Run("DeleteAddress", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user and profile
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		profile := tests.CreateTestProfile(db, user.ID)

		// Create an address
		address := tests.CreateTestAddress(db, profile.ID, true)

		// Delete the address
		err := db.Delete(&address).Error
		assert.Nil(t, err, "Expected no error when deleting address")

		// Try to retrieve the deleted address
		var retrievedAddress models.Address
		err = db.Where("id = ?", address.ID).First(&retrievedAddress).Error
		assert.Error(t, err, "Expected error when retrieving deleted address")
	})

	t.Run("MultipleAddresses_WithDefault", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user and profile
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		profile := tests.CreateTestProfile(db, user.ID)

		// Create a default address
		tests.CreateTestAddress(db, profile.ID, true)

		// Create a second non-default address
		address2 := models.Address{
			UserProfileID: fmt.Sprintf("%d", profile.ID),
			Name:          "Work",
			Street:        "789 Work St",
			City:          "Work City",
			State:         "Work State",
			ZipCode:       "67890",
			Country:       "USA",
			IsDefault:     false,
		}

		err := db.Create(&address2).Error
		assert.Nil(t, err, "Expected no error when creating second address")

		// Verify profile has two addresses and address1 is the default
		var retrievedProfile models.UserProfile
		err = db.Preload("Addresses").First(&retrievedProfile, profile.ID).Error
		assert.Nil(t, err, "Expected no error when retrieving profile with addresses")
		assert.Len(t, retrievedProfile.Addresses, 2, "Expected profile to have 2 addresses")

		// Verify address1 is marked as default
		var defaultAddresses int64
		err = db.Model(&models.Address{}).Where("user_profile_id = ? AND is_default = ?", profile.ID, true).Count(&defaultAddresses).Error
		assert.Nil(t, err, "Expected no error when counting default addresses")
		assert.Equal(t, int64(1), defaultAddresses, "Expected exactly one default address")
		var addresses []models.Address
		err = db.Where("user_profile_id = ?", profile.ID).Find(&addresses).Error
		assert.Nil(t, err, "Expected no error when retrieving addresses")
		assert.Equal(t, 2, len(addresses), "Expected profile to have two addresses")

		// Verify only one address is default
		defaultCount := 0
		for _, addr := range addresses {
			if addr.IsDefault {
				defaultCount++
			}
		}
		assert.Equal(t, 1, defaultCount, "Expected exactly one default address")

		// Set the second address as default
		address2.IsDefault = true
		err = db.Save(&address2).Error
		assert.Nil(t, err, "Expected no error when setting new default address")

		// Update the profile to use the new default address
		db.Model(&models.UserProfile{}).Where("id = ?", profile.ID).Update("default_address_id", address2.ID)

		// Retrieve the updated profile
		var updatedProfile models.UserProfile
		db.Preload("Addresses").Where("id = ?", profile.ID).First(&updatedProfile)

		// Verify new default address is set correctly in the profile
		assert.Equal(t, address2.ID, updatedProfile.DefaultAddressID)
	})
}
