package testutils

import (
	"meals/config"
	"meals/store"

	"gorm.io/gorm"
)

// SetupTestDB creates a test PostgreSQL database for testing
func SetupTestDB() *gorm.DB {
	config.InitConfig()
	store.InitDB()
	return store.DB
}

// CleanupTestDB removes all records from test tables
func CleanupTestDB(db *gorm.DB) error {
	// Delete all records from all tables in reverse order to avoid FK conflicts
	tables, err := store.DB.Migrator().GetTables()
	if err != nil {
		return err
	}
	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}
	return nil
}
