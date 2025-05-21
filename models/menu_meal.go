package models

import (
	"gorm.io/gorm"
)

type MenuMeal struct {
	gorm.Model
	DeliveryDay string `json:"delivery_day" gorm:"type:varchar(20);not null"`
	MenuID      uint   `json:"menu_id" gorm:"not null"`                                          // Foreign key to Menu
	MealID      uint   `json:"meal_id" gorm:"not null"`                                          // Foreign key to Meal
	Menu        Menu   `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"`  // Reference to Menu
	Meal        Meal   `gorm:"foreignKey:MealID;constraint:OnDelete:RESTRICT;OnUpdate:CASCADE;"` // Reference to Meal
}
