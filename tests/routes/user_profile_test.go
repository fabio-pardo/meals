package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"meals/handlers"
	"meals/models"
	"meals/tests"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Helper to authenticate a user in the test context
func authenticateUser(c *gin.Context, user models.User) {
	c.Set("user", user)
}

func TestUserProfileEndpoints(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("GetUserProfile_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user and profile
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		profile := tests.CreateTestProfile(db, user.ID)
		address := tests.CreateTestAddress(db, profile.ID, true)

		// Update the profile with default address
		db.Model(&profile).Update("default_address_id", address.ID)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/profiles/me", nil)

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		authenticateUser(c, user)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userID := user.ID
			var userProfile models.UserProfile

			result := db.Where("user_id = ?", userID).Preload("Addresses").First(&userProfile)
			if result.Error != nil {
				handlers.RespondWithError(c, handlers.NotFoundError("User profile not found"))
				return
			}

			c.JSON(http.StatusOK, userProfile)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response body
		var responseProfile models.UserProfile
		err := json.Unmarshal(w.Body.Bytes(), &responseProfile)
		assert.Nil(t, err)
		assert.Equal(t, profile.ID, responseProfile.ID)
		assert.Equal(t, user.ID, responseProfile.UserID)
		assert.NotEmpty(t, responseProfile.Addresses)
	})

	t.Run("UpdateUserProfile_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user and profile
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Prepare update data
		updateData := map[string]interface{}{
			"phone_number":        "555-987-6543",
			"dietary_preferences": `{"vegan": true, "allergies": ["gluten"]}`,
		}

		jsonData, _ := json.Marshal(updateData)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/profiles/me", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Add user authentication to context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		authenticateUser(c, user)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userID := user.ID
			var updateRequest struct {
				PhoneNumber        string   `json:"phone_number"`
				DietaryPreferences []string `json:"dietary_preferences"`
			}

			if err := c.ShouldBindJSON(&updateRequest); err != nil {
				handlers.RespondWithError(c, handlers.ValidationError("input", "Invalid input data"))
				return
			}

			var userProfile models.UserProfile
			result := db.Where("user_id = ?", userID).First(&userProfile)
			if result.Error != nil {
				// Create a new profile if none exists
				userProfile = models.UserProfile{
					UserID:             userID,
					PhoneNumber:        updateRequest.PhoneNumber,
					DietaryPreferences: updateRequest.DietaryPreferences,
				}
				db.Create(&userProfile)
			} else {
				// Update existing profile
				userProfile.PhoneNumber = updateRequest.PhoneNumber
				userProfile.DietaryPreferences = updateRequest.DietaryPreferences
				db.Save(&userProfile)
			}

			c.JSON(http.StatusOK, userProfile)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response body
		var responseProfile models.UserProfile
		err := json.Unmarshal(w.Body.Bytes(), &responseProfile)
		assert.Nil(t, err)
		assert.Equal(t, "555-987-6543", responseProfile.PhoneNumber)
		assert.Equal(t, `{"vegan": true, "allergies": ["gluten"]}`, responseProfile.DietaryPreferences)
	})
}

func TestAddressEndpoints(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("CreateAddress_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Prepare new address data
		addressData := models.Address{
			Name:      "Office",
			Street:    "456 Office St",
			City:      "Office City",
			State:     "Office State",
			ZipCode:   "67890",
			Country:   "USA",
			IsDefault: true,
		}

		jsonData, _ := json.Marshal(addressData)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/profiles/me/addresses", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Add user authentication to context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		authenticateUser(c, user)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			var userProfile models.UserProfile
			result := db.Where("user_id = ?", user.ID).First(&userProfile)
			if result.Error != nil {
				handlers.RespondWithError(c, handlers.NotFoundError("User profile not found"))
				return
			}

			var address models.Address
			if err := c.ShouldBindJSON(&address); err != nil {
				handlers.RespondWithError(c, handlers.ValidationError("input", "Invalid address data"))
				return
			}

			address.UserProfileID = userProfile.ID

			if err := db.Create(&address).Error; err != nil {
				handlers.RespondWithError(c, handlers.DatabaseError("Failed to create address"))
				return
			}

			// If this is the default address, update the profile
			if address.IsDefault {
				// If this is the first address or set as default, update the profile's default address
				db.Model(&userProfile).Update("default_address_id", address.ID)

				// If there are other addresses, make sure they are not default
				if address.IsDefault {
					db.Model(&models.Address{}).
						Where("user_profile_id = ? AND id != ?", userProfile.ID, address.ID).
						Update("is_default", false)
				}
			}

			c.JSON(http.StatusCreated, address)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify response body
		var responseAddress models.Address
		err := json.Unmarshal(w.Body.Bytes(), &responseAddress)
		assert.Nil(t, err)
		assert.Equal(t, addressData.Name, responseAddress.Name)
		assert.Equal(t, addressData.Street, responseAddress.Street)
		assert.Equal(t, user.ID, responseAddress.UserProfileID)
		assert.True(t, responseAddress.IsDefault)

		// Verify profile's default address was updated
		var updatedProfile models.UserProfile
		db.First(&updatedProfile, user.ID)
		assert.Equal(t, responseAddress.ID, updatedProfile.DefaultAddressID)
	})

	t.Run("DeleteAddress_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create two addresses
		address1 := tests.CreateTestAddress(db, user.ID, true)
		address2 := tests.CreateTestAddress(db, user.ID, false)
		address2.Name = "Work"
		db.Save(&address2)

		// Update profile's default address
		db.Model(&user).Update("default_address_id", address1.ID)

		// Create a test request to delete the default address
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/profiles/me/addresses/%d", address1.ID), nil)

		// Add user authentication to context
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Set user authentication
		authenticateUser(c, user)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", address1.ID)}}

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			addressID := address1.ID

			var address models.Address
			if err := db.First(&address, addressID).Error; err != nil {
				handlers.RespondWithError(c, handlers.NotFoundError("Address not found"))
				return
			}

			// Get the profile to check if this is the default address
			var profile models.UserProfile
			if err := db.First(&profile, address.UserProfileID).Error; err != nil {
				handlers.RespondWithError(c, handlers.DatabaseError("Failed to retrieve profile"))
				return
			}

			// If this is the default address, find another address to make default
			if profile.DefaultAddressID == &address.ID {
				var newDefaultAddress models.Address
				err := db.Where("user_profile_id = ? AND id != ?", profile.ID, address.ID).
					First(&newDefaultAddress).Error

				if err == nil {
					// Set the new address as default
					newDefaultAddress.IsDefault = true
					db.Save(&newDefaultAddress)

					// Update profile's default address
					profile.DefaultAddressID = &newDefaultAddress.ID
					db.Save(&profile)
				} else {
					// No other address found, reset default address
					profile.DefaultAddressID = nil
					db.Save(&profile)
				}
			}

			// Delete the address
			if err := db.Delete(&address).Error; err != nil {
				handlers.RespondWithError(c, handlers.DatabaseError("Failed to delete address"))
				return
			}

			c.Status(http.StatusNoContent)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify address was deleted
		var deletedAddress models.Address
		err := db.First(&deletedAddress, address1.ID).Error
		assert.Error(t, err, "Expected address to be deleted")

		// Verify profile's default address was updated to address2
		var updatedProfile models.UserProfile
		db.First(&updatedProfile, user.ID)
		assert.Equal(t, address2.ID, updatedProfile.DefaultAddressID)
	})
}
