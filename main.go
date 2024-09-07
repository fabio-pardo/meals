package main

import (
	"log"
	"meals/config"
	"meals/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file")
	}
}

func main() {
	// Initialize the database
	config.InitDB()

	// Set up the router
	router := gin.Default()

	// Register routes
	routes.RegisterRoutes(router)

	// Run the server
	router.Run("localhost:8080")
}
