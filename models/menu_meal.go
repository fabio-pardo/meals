package models

import "time"

type MenuMeal struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt   time.Time `gorm:"autoCreateTime;not null"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;not null"`
	DeliveryDay string    `json:"delivery_day" gorm:"type:varchar(20);not null"`
	MenuID      uint      `gorm:"not null"`                                                        // Foreign key to Menu
	MealID      uint      `json:"meal_id" gorm:"not null"`                                         // Foreign key to Meal
	Menu        Menu      `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"` // Reference to Menu
	Meal        Meal      `gorm:"foreignKey:MealID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"` // Reference to Meal
}
