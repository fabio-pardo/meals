package models

import "time"

type Menu struct {
	ID            uint       `gorm:"primaryKey;autoIncrement;not null"`
	WeekStartDate time.Time  `gorm:"not null"`
	WeekEndDate   time.Time  `gorm:"not null"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;not null"`
	MenuMeals     []MenuMeal `gorm:"foreignKey:MenuID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
}
