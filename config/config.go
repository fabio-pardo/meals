package config

import (
	"log"

	"github.com/joho/godotenv"
)

func InitConfig() {
	// Load dotenv
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file")
	}
}
