package models

import "time"

type Menu struct {
	ID            uint       `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name          string     `json:"name" gorm:"not null"`
	Description   string     `json:"description"`
	WeekStartDate time.Time  `json:"week_start_date" gorm:"not null"`
	WeekEndDate   time.Time  `json:"week_end_date" gorm:"not null"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;not null"`
	MenuMeals     []MenuMeal `json:"menu_meals" gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"`
	MealIDs       []uint     `json:"meal_ids" gorm:"-"` // Used for handling many-to-many relationships, not stored in DB
}
