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
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the DB: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	if err := DB.AutoMigrate(&models.Meal{}, &models.Menu{}, &models.MenuMeal{}, &models.User{}); err != nil {
		log.Fatalf("Failed to migrate the DB: %v", err)
	}

	log.Println("Migrated PostgreSQL DB successfully")
}
