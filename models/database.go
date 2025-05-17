package models

import (
	"gorm.io/gorm"
)

// Database is a wrapper for gorm.DB to be used in context
type Database struct {
	DB *gorm.DB
}

// GetDB returns the underlying gorm.DB
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}
