package store

import (
	"fmt"
	"log"
	"meals/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = ""
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "meals"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

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
