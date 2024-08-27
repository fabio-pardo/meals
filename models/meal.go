package models

import "time"

type Meal struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Price     float64   `json:"price" gorm:"not null"`
}
