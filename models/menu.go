package models

import "time"

type Menu struct {
	ID            uint       `gorm:"primaryKey;autoIncrement;not null"`
	WeekStartDate time.Time  `json:"week_start_date" gorm:"not null"`
	WeekEndDate   time.Time  `json:"week_end_date" gorm:"not null"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;not null"`
	MenuMeals     []MenuMeal `json:"menu_meals" gorm:"foreignKey:MenuID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL;"`
}
