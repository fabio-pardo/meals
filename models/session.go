package models

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a user authentication session
type Session struct {
	gorm.Model
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	UserIdentifier string    `json:"user_id" gorm:"column:user_identifier;type:varchar(50);not null"`
	User           *User     `json:"user" gorm:"foreignKey:UserIdentifier;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SetExpiration sets the expiration time for a session
func (s *Session) SetExpiration(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}
