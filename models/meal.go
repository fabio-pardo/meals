package models

import (
	"gorm.io/gorm"
)

type Meal struct {
	gorm.Model
	Name  string  `json:"name" gorm:"size:255;not null"`
	Price float64 `json:"price" gorm:"not null"`
}
