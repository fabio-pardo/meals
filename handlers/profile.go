package handlers

import (
	"encoding/json"
	"meals/models"
	"meals/store"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProfileResponse represents the response for profile operations
type ProfileResponse struct {
	Profile models.UserProfile `json:"profile"`
}

// AddressResponse represents the response for address operations
type AddressResponse struct {
	Address models.Address `json:"address"`
}

// CreateProfileRequest represents the request body for creating a user profile
type CreateProfileRequest struct {
	PhoneNumber        string   `json:"phone_number"`
	DietaryPreferences []string `json:"dietary_preferences"`
}

// UpdateProfileRequest represents the request body for updating a user profile
type UpdateProfileRequest struct {
	PhoneNumber        string   `json:"phone_number"`
	DefaultAddressID   *uint    `json:"default_address_id"`
	DietaryPreferences []string `json:"dietary_preferences"`
}

// CreateAddressRequest represents the request body for creating an address
type CreateAddressRequest struct {
	Name         string   `json:"name" binding:"required"`
	Street       string   `json:"street" binding:"required"`
	Unit         string   `json:"unit"`
	City         string   `json:"city" binding:"required"`
	State        string   `json:"state" binding:"required"`
	ZipCode      string   `json:"zip_code" binding:"required"`
	Country      string   `json:"country"`
	IsDefault    bool     `json:"is_default"`
	Instructions string   `json:"instructions"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

// GetUserProfileHandler handles fetching the profile of the authenticated user
func GetUserProfileHandler(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	var profile models.UserProfile
	db := store.GetTxFromContext(c)

	// Find profile with preloaded addresses
	result := db.Where("user_id = ?", userID).Preload("Addresses").First(&profile)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// If no profile exists, return empty but valid response
			c.JSON(http.StatusOK, ProfileResponse{
				Profile: models.UserProfile{
					UserID: userID.(uint),
				},
			})
			return
		}

		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch profile",
			Details: result.Error.Error(),
		})
		return
	}

	// Unmarshal preferences JSON if exists
	if profile.PreferencesJSON != "" {
		if err := json.Unmarshal([]byte(profile.PreferencesJSON), &profile.DietaryPreferences); err != nil {
			// Don't fail the request, just log the error and continue
			profile.DietaryPreferences = []string{}
		}
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Profile: profile,
	})
}

// CreateOrUpdateProfileHandler handles creating or updating a user's profile
func CreateOrUpdateProfileHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, ValidationError("Invalid request data", err.Error()))
		return
	}

	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// Check if profile exists
		var profile models.UserProfile
		result := tx.Where("user_id = ?", userID).First(&profile)

		isNew := false
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Create new profile
				profile = models.UserProfile{
					UserID: userID.(uint),
				}
				isNew = true
			} else {
				return DatabaseErrorType{
					Message: "Failed to check for existing profile",
					Details: result.Error.Error(),
				}
			}
		}

		// Update profile fields
		profile.PhoneNumber = req.PhoneNumber
		profile.DefaultAddressID = req.DefaultAddressID

		// Marshal dietary preferences to JSON
		if req.DietaryPreferences != nil {
			preferencesJSON, err := json.Marshal(req.DietaryPreferences)
			if err != nil {
				return ValidationErrorType{
					Message: "Failed to marshal dietary preferences",
					Details: err.Error(),
				}
			}
			profile.PreferencesJSON = string(preferencesJSON)
		}

		// Save profile
		if isNew {
			if err := tx.Create(&profile).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to create profile",
					Details: err.Error(),
				}
			}
		} else {
			if err := tx.Save(&profile).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update profile",
					Details: err.Error(),
				}
			}
		}

		// Update default address if provided
		if req.DefaultAddressID != nil && *req.DefaultAddressID > 0 {
			// First, reset all default flags for this user's addresses
			if err := tx.Model(&models.Address{}).
				Where("user_profile_id = ?", profile.ID).
				Update("is_default", false).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update address defaults",
					Details: err.Error(),
				}
			}

			// Then set the new default
			if err := tx.Model(&models.Address{}).
				Where("id = ? AND user_profile_id = ?", *req.DefaultAddressID, profile.ID).
				Update("is_default", true).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to set default address",
					Details: err.Error(),
				}
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	// Fetch and return the updated profile
	var updatedProfile models.UserProfile
	if err := store.DB.Where("user_id = ?", userID).Preload("Addresses").First(&updatedProfile).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch updated profile",
			Details: err.Error(),
		})
		return
	}

	// Unmarshal preferences JSON
	if updatedProfile.PreferencesJSON != "" {
		if err := json.Unmarshal([]byte(updatedProfile.PreferencesJSON), &updatedProfile.DietaryPreferences); err != nil {
			updatedProfile.DietaryPreferences = []string{}
		}
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Profile: updatedProfile,
	})
}

// ListAddressesHandler handles fetching all addresses for the authenticated user
func ListAddressesHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	var profile models.UserProfile
	db := store.GetTxFromContext(c)

	// First get the user's profile ID
	result := db.Where("user_id = ?", userID).First(&profile)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// If no profile exists, return empty addresses array
			c.JSON(http.StatusOK, gin.H{
				"addresses": []models.Address{},
			})
			return
		}

		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch profile",
			Details: result.Error.Error(),
		})
		return
	}

	// Now get the addresses
	var addresses []models.Address
	if err := db.Where("user_profile_id = ?", profile.ID).Find(&addresses).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch addresses",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"addresses": addresses,
	})
}

// CreateAddressHandler handles adding a new address
func CreateAddressHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	var req CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, ValidationError("Invalid request data", err.Error()))
		return
	}

	var address models.Address
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// First get or create user profile
		var profile models.UserProfile
		result := tx.Where("user_id = ?", userID).First(&profile)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Create new profile
				profile = models.UserProfile{
					UserID: userID.(uint),
				}
				if err := tx.Create(&profile).Error; err != nil {
					return DatabaseErrorType{
						Message: "Failed to create profile",
						Details: err.Error(),
					}
				}
			} else {
				return DatabaseErrorType{
					Message: "Failed to check for existing profile",
					Details: result.Error.Error(),
				}
			}
		}

		// Create address
		address = models.Address{
			UserProfileID: profile.ID,
			Name:          req.Name,
			Street:        req.Street,
			Unit:          req.Unit,
			City:          req.City,
			State:         req.State,
			ZipCode:       req.ZipCode,
			Country:       req.Country,
			IsDefault:     req.IsDefault,
			Instructions:  req.Instructions,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
		}

		// Validate address
		if validationErrors := address.ValidateAddress(); len(validationErrors) > 0 {
			return ValidationErrorType{
				Message: "Address validation failed",
				Details: validationErrors,
			}
		}

		// If this is set as default address
		if address.IsDefault {
			// Reset all other default flags
			if err := tx.Model(&models.Address{}).
				Where("user_profile_id = ?", profile.ID).
				Update("is_default", false).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update address defaults",
					Details: err.Error(),
				}
			}
		}

		// Save address
		if err := tx.Create(&address).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to create address",
				Details: err.Error(),
			}
		}

		// Update profile's default address if this is default or first address
		var addressCount int64
		tx.Model(&models.Address{}).Where("user_profile_id = ?", profile.ID).Count(&addressCount)

		if address.IsDefault || addressCount == 1 {
			profile.DefaultAddressID = &address.ID
			if err := tx.Save(&profile).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update profile's default address",
					Details: err.Error(),
				}
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusCreated, AddressResponse{
		Address: address,
	})
}

// GetAddressHandler handles fetching a specific address
func GetAddressHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid address ID"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	var address models.Address
	db := store.GetTxFromContext(c)

	// First get the user's profile ID
	var profile models.UserProfile
	result := db.Where("user_id = ?", userID).First(&profile)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			HandleAppError(c, NotFoundErrorType{Resource: "Profile"})
			return
		}

		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch profile",
			Details: result.Error.Error(),
		})
		return
	}

	// Now get the address
	result = db.Where("id = ? AND user_profile_id = ?", id, profile.ID).First(&address)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			HandleAppError(c, NotFoundErrorType{Resource: "Address"})
			return
		}

		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch address",
			Details: result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AddressResponse{
		Address: address,
	})
}

// UpdateAddressHandler handles updating an address
func UpdateAddressHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid address ID"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	var req CreateAddressRequest // Reuse the same request structure
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, ValidationError("Invalid request data", err.Error()))
		return
	}

	err = store.WithTransaction(c, func(tx *gorm.DB) error {
		// Get the user's profile
		var profile models.UserProfile
		if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Profile"}
			}
			return DatabaseErrorType{
				Message: "Failed to fetch profile",
				Details: err.Error(),
			}
		}

		// Get the address
		var address models.Address
		if err := tx.Where("id = ? AND user_profile_id = ?", id, profile.ID).First(&address).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Address"}
			}
			return DatabaseErrorType{
				Message: "Failed to fetch address",
				Details: err.Error(),
			}
		}

		// Update address fields
		address.Name = req.Name
		address.Street = req.Street
		address.Unit = req.Unit
		address.City = req.City
		address.State = req.State
		address.ZipCode = req.ZipCode
		address.Country = req.Country
		address.IsDefault = req.IsDefault
		address.Instructions = req.Instructions
		address.Latitude = req.Latitude
		address.Longitude = req.Longitude

		// Validate address
		if validationErrors := address.ValidateAddress(); len(validationErrors) > 0 {
			return ValidationErrorType{
				Message: "Address validation failed",
				Details: validationErrors,
			}
		}

		// If this is set as default address
		if address.IsDefault {
			// Reset all other default flags
			if err := tx.Model(&models.Address{}).
				Where("user_profile_id = ? AND id != ?", profile.ID, address.ID).
				Update("is_default", false).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update address defaults",
					Details: err.Error(),
				}
			}

			// Update profile's default address
			profile.DefaultAddressID = &address.ID
			if err := tx.Save(&profile).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update profile's default address",
					Details: err.Error(),
				}
			}
		}

		// Save address
		if err := tx.Save(&address).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to update address",
				Details: err.Error(),
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	// Get updated address
	var updatedAddress models.Address
	if err := store.DB.First(&updatedAddress, id).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch updated address",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AddressResponse{
		Address: updatedAddress,
	})
}

// DeleteAddressHandler handles deleting an address
func DeleteAddressHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid address ID"))
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	err = store.WithTransaction(c, func(tx *gorm.DB) error {
		// Get the user's profile
		var profile models.UserProfile
		if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Profile"}
			}
			return DatabaseErrorType{
				Message: "Failed to fetch profile",
				Details: err.Error(),
			}
		}

		// Find address to verify it belongs to this user
		var address models.Address
		if err := tx.Where("id = ? AND user_profile_id = ?", id, profile.ID).First(&address).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Address"}
			}
			return DatabaseErrorType{
				Message: "Failed to fetch address",
				Details: err.Error(),
			}
		}

		// If this is the default address, we need to handle that
		if address.IsDefault {
			// Find another address to make default
			var newDefaultAddress models.Address
			if err := tx.Where("user_profile_id = ? AND id != ?", profile.ID, address.ID).
				Order("created_at desc").First(&newDefaultAddress).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return DatabaseErrorType{
						Message: "Failed to fetch alternate default address",
						Details: err.Error(),
					}
				}

				// If no other address, set default address ID to nil
				profile.DefaultAddressID = nil
			} else {
				// Make the newer address the default
				newDefaultAddress.IsDefault = true
				if err := tx.Save(&newDefaultAddress).Error; err != nil {
					return DatabaseErrorType{
						Message: "Failed to update new default address",
						Details: err.Error(),
					}
				}

				profile.DefaultAddressID = &newDefaultAddress.ID
			}

			// Update profile
			if err := tx.Save(&profile).Error; err != nil {
				return DatabaseErrorType{
					Message: "Failed to update profile's default address",
					Details: err.Error(),
				}
			}
		}

		// Delete the address
		if err := tx.Delete(&address).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to delete address",
				Details: err.Error(),
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Address deleted successfully",
	})
}

// SetDriverProfileHandler handles updating driver-specific profile data
func SetDriverProfileHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "Authentication required",
		})
		return
	}

	// This endpoint is only for drivers
	userType, ok := c.Get("userType")
	if !ok || (userType != models.UserTypeDriver && userType != models.UserTypeAdmin) {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusForbidden,
			Code:    ErrForbidden,
			Message: "Only drivers can update driver profile data",
		})
		return
	}

	var req struct {
		IsAvailable   *bool   `json:"is_available"`
		VehicleType   *string `json:"vehicle_type"`
		LicenseNumber *string `json:"license_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, ValidationError("Invalid request data", err.Error()))
		return
	}

	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// Get or create profile
		var profile models.UserProfile
		result := tx.Where("user_id = ?", userID).First(&profile)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Create new profile
				profile = models.UserProfile{
					UserID: userID.(uint),
				}
				if err := tx.Create(&profile).Error; err != nil {
					return DatabaseErrorType{
						Message: "Failed to create profile",
						Details: err.Error(),
					}
				}
			} else {
				return DatabaseErrorType{
					Message: "Failed to check for existing profile",
					Details: result.Error.Error(),
				}
			}
		}

		// Update driver-specific fields
		if req.IsAvailable != nil {
			profile.IsAvailable = req.IsAvailable
		}
		if req.VehicleType != nil {
			profile.VehicleType = req.VehicleType
		}
		if req.LicenseNumber != nil {
			profile.LicenseNumber = req.LicenseNumber
		}

		// Save profile
		if err := tx.Save(&profile).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to update driver profile",
				Details: err.Error(),
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	// Get updated profile
	var updatedProfile models.UserProfile
	if err := store.DB.Where("user_id = ?", userID).First(&updatedProfile).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch updated profile",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		Profile: updatedProfile,
	})
}
