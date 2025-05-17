package models

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a user authentication session
type Session struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	User      User      `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SetExpiration sets the expiration time for a session
func (s *Session) SetExpiration(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}
