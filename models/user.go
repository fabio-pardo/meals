package models

import (
	"time"

	"github.com/markbates/goth"
	"gorm.io/gorm"
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
	UserID            string    `json:"user_id" gorm:"unique;not null"`
	AccessToken       string    `json:"access_token" gorm:"not null"`
	AccessTokenSecret string    `json:"access_token_secret"`
	RefreshToken      string    `json:"refresh_token"`
	ExpiresAt         time.Time `json:"expires_at" gorm:"not null"`
	IDToken           string    `json:"id_token" gorm:"not null"`
}

func ConvertGothUserToModelUser(gothUser *goth.User) (*User, error) {
	var user User
	user.Provider = gothUser.Provider
	user.UserID = gothUser.UserID
	user.Name = gothUser.Name
	user.Email = gothUser.Email
	user.AccessToken = gothUser.AccessToken
	user.AccessTokenSecret = gothUser.AccessTokenSecret
	user.RefreshToken = gothUser.RefreshToken
	user.ExpiresAt = gothUser.ExpiresAt
	return &user, nil
}
