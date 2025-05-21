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

// AfterDelete hook ensures that MenuMeals are soft deleted when a Menu is soft deleted
func (m *Menu) AfterDelete(tx *gorm.DB) error {
	// This is more efficient than the BeforeDelete approach as it uses a single UPDATE
	// statement rather than individual DELETE calls for each record.
	// It also ensures the deleted_at timestamp matches the parent record.
	return tx.Model(&MenuMeal{}).Where("menu_id = ?", m.ID).Update("deleted_at", m.DeletedAt).Error
}
