package models

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	gorm.Model
	Name          string     `json:"name" gorm:"not null"`
	Description   string     `json:"description"`
	WeekStartDate time.Time  `json:"week_start_date" gorm:"not null"`
	WeekEndDate   time.Time  `json:"week_end_date" gorm:"not null"`
	MenuMeals     []MenuMeal `json:"menu_meals" gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"`
}
