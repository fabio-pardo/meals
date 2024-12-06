package store

import (
	"log"
	"meals/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
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
