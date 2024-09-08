package main

import (
	"log"
	"meals/auth"
	"meals/config"
	"meals/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load dotenv
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file")
	}

	// Init OAuth2
	auth.InitOAuth2()

	// Initialize the database
	config.InitDB()

	// Set up the router
	router := gin.Default()

	// Register routes
	routes.RegisterRoutes(router)

	// Run the server
	router.Run("localhost:8080")
}
