package models

import (
	"gorm.io/gorm"
)

// UserProfile stores additional information about users, including delivery addresses
type UserProfile struct {
	gorm.Model
	UserID uint
	User   User
}
