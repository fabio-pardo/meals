package store

import (
	"log"
	"meals/config"
	"meals/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Use the configuration from the config package
	dbConfig := config.AppConfig.Database

	// Get the DSN from our config
	dsn := dbConfig.GetDSN()

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // Disable foreign key checks during migration
	})
	if err != nil {
		log.Fatalf("Failed to connect to the DB: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	// First migrate models without relationships or with simple relationships
	if err := DB.AutoMigrate(
		&models.Meal{},
		&models.User{},
	); err != nil {
		log.Fatalf("Failed to migrate basic models: %v", err)
	}

	// Then migrate models with relationships but not circular ones
	if err := DB.AutoMigrate(
		&models.Menu{},
		&models.UserProfile{},
	); err != nil {
		log.Fatalf("Failed to migrate relationship models: %v", err)
	}

	// Finally migrate models with complex relationships
	if err := DB.AutoMigrate(
		&models.MenuMeal{},
		&models.Address{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		log.Fatalf("Failed to migrate complex relationship models: %v", err)
	}

	// Enable foreign key constraints after migration
	if err := DB.Exec("SET session_replication_role = 'origin';").Error; err != nil {
		log.Fatalf("Failed to re-enable foreign key constraints: %v", err)
	}

	log.Println("Migrated PostgreSQL DB successfully")
}
