package handlers

import (
	"meals/models"
	"meals/store"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProfileResponse represents the response for profile operations
type ProfileResponse struct {
	Profile models.UserProfile `json:"profile"`
}

// CreateProfileRequest represents the request body for creating a user profile
type CreateProfileRequest struct {
}

// UpdateProfileRequest represents the request body for updating a user profile
type UpdateProfileRequest struct {
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

	// Find profile
	result := db.Where("user_id = ?", userID).First(&profile)
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

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	// Fetch and return the updated profile
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
				"addresses": []gin.H{},
			})
			return
		}

		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch profile",
			Details: result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"addresses": []gin.H{},
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

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"address": gin.H{},
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
