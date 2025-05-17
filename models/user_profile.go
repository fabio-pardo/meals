package models

import (
	"encoding/json"
	"time"
	"gorm.io/gorm"
)

// UserProfile stores additional information about users, including delivery addresses
type UserProfile struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	UserID           uint           `json:"user_id" gorm:"not null;uniqueIndex"`
	PhoneNumber      string         `json:"phone_number" gorm:"type:varchar(20)"`
	DefaultAddressID *uint          `json:"default_address_id" gorm:""`
	Addresses        []Address      `json:"addresses" gorm:"foreignKey:UserProfileID;constraint:OnDelete:CASCADE;"`

	// Driver-specific fields
	IsAvailable   *bool   `json:"is_available,omitempty" gorm:"default:false"`
	VehicleType   *string `json:"vehicle_type,omitempty" gorm:"type:varchar(50)"`
	LicenseNumber *string `json:"license_number,omitempty" gorm:"type:varchar(50)"`

	// Customer preferences
	DietaryPreferences []string `json:"dietary_preferences" gorm:"-"` // Stored as JSON in PreferencesJSON
	PreferencesJSON    string   `json:"preferences_json" gorm:"type:text"`
}

// BeforeSave is a GORM hook that runs before saving the UserProfile
func (up *UserProfile) BeforeSave(tx *gorm.DB) error {
	// Convert DietaryPreferences to JSON and store in PreferencesJSON
	if up.DietaryPreferences != nil {
		jsonData, err := json.Marshal(up.DietaryPreferences)
		if err != nil {
			return err
		}
		up.PreferencesJSON = string(jsonData)
	} else {
		up.PreferencesJSON = "[]"
	}
	return nil
}

// AfterFind is a GORM hook that runs after finding a UserProfile
func (up *UserProfile) AfterFind(tx *gorm.DB) error {
	// Convert PreferencesJSON back to DietaryPreferences
	if up.PreferencesJSON != "" {
		var prefs []string
		if err := json.Unmarshal([]byte(up.PreferencesJSON), &prefs); err != nil {
			return err
		}
		up.DietaryPreferences = prefs
	}
	return nil
}

// Address represents a delivery address associated with a user
type Address struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	UserProfileID uint           `json:"user_profile_id" gorm:"not null"`
	Name          string         `json:"name" gorm:"type:varchar(100);not null"` // e.g., "Home", "Work"
	Street        string         `json:"street" gorm:"type:varchar(255);not null"`
	Unit          string         `json:"unit" gorm:"type:varchar(50)"`
	City          string         `json:"city" gorm:"type:varchar(100);not null"`
	State         string         `json:"state" gorm:"type:varchar(100);not null"`
	ZipCode       string         `json:"zip_code" gorm:"type:varchar(20);not null"`
	Country       string         `json:"country" gorm:"type:varchar(100);not null;default:'USA'"`
	IsDefault     bool           `json:"is_default" gorm:"default:false"`
	Instructions  string         `json:"instructions" gorm:"type:text"`
	Latitude      *float64       `json:"latitude" gorm:"type:decimal(10,7)"`
	Longitude     *float64       `json:"longitude" gorm:"type:decimal(10,7)"`
}

// ValidateAddress validates address data
func (a *Address) ValidateAddress() []string {
	var errors []string

	if a.Name == "" {
		errors = append(errors, "Address name is required")
	}

	if a.Street == "" {
		errors = append(errors, "Street is required")
	}

	if a.City == "" {
		errors = append(errors, "City is required")
	}

	if a.State == "" {
		errors = append(errors, "State is required")
	}

	if a.ZipCode == "" {
		errors = append(errors, "ZIP code is required")
	}

	return errors
}

// GetFormattedAddress returns a formatted address string
func (a *Address) GetFormattedAddress() string {
	formatted := a.Street
	if a.Unit != "" {
		formatted += ", " + a.Unit
	}
	formatted += ", " + a.City + ", " + a.State + " " + a.ZipCode
	if a.Country != "USA" && a.Country != "" {
		formatted += ", " + a.Country
	}
	return formatted
}
