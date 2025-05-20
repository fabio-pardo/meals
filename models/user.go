package models

import (
	"time"

	"github.com/markbates/goth"
	"gorm.io/gorm"
)

// UserType represents the role of a user in the system
type UserType string

const (
	UserTypeAdmin    UserType = "admin"
	UserTypeDriver   UserType = "driver"
	UserTypeCustomer UserType = "customer"
)

type User struct {
	gorm.Model
	Provider          string    `json:"provider" gorm:"not null"`
	Email             string    `json:"email" gorm:"unique;not null"`
	Name              string    `json:"name"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	NickName          string    `json:"nickname"`
	Description       string    `json:"description"`
	AccessToken       string    `json:"access_token" gorm:"not null"`
	AccessTokenSecret string    `json:"access_token_secret"`
	RefreshToken      string    `json:"refresh_token"`
	ExpiresAt         time.Time `json:"expires_at" gorm:"not null"`
	IDToken           string    `json:"id_token" gorm:"not null"`
	UserID            string    `json:"user_id" gorm:"type:varchar(50);unique;not null"`
	UserType          UserType  `json:"user_type" gorm:"type:varchar(20);default:'customer'"`
}

// ValidateUser validates the user data
func (u *User) ValidateUser() []string {
	var errors []string

	if u.Email == "" {
		errors = append(errors, "Email is required")
	}

	if u.UserID == "" {
		errors = append(errors, "UserID is required")
	}

	if u.Provider == "" {
		errors = append(errors, "Provider is required")
	}

	// Check if the user type is valid
	if u.UserType != "" &&
		u.UserType != UserTypeAdmin &&
		u.UserType != UserTypeDriver &&
		u.UserType != UserTypeCustomer {
		errors = append(errors, "Invalid user type")
	}

	return errors
}

// IsAdmin checks if the user has admin privileges
func (u *User) IsAdmin() bool {
	return u.UserType == UserTypeAdmin
}

// IsDriver checks if the user is a driver
func (u *User) IsDriver() bool {
	return u.UserType == UserTypeDriver
}

// IsCustomer checks if the user is a customer
func (u *User) IsCustomer() bool {
	return u.UserType == UserTypeCustomer || u.UserType == ""
}

func ConvertGothUserToModelUser(gothUser *goth.User) (*User, error) {
	var user User
	user.Provider = gothUser.Provider
	user.UserID = gothUser.UserID
	user.Name = gothUser.Name
	user.Email = gothUser.Email
	user.FirstName = gothUser.FirstName
	user.LastName = gothUser.LastName
	user.NickName = gothUser.NickName
	user.Description = gothUser.Description
	user.AccessToken = gothUser.AccessToken
	user.AccessTokenSecret = gothUser.AccessTokenSecret
	user.RefreshToken = gothUser.RefreshToken
	user.ExpiresAt = gothUser.ExpiresAt
	user.IDToken = gothUser.IDToken
	// Default new users to customers
	user.UserType = UserTypeCustomer

	return &user, nil
}
