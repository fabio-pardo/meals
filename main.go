package main

import (
	"meals/config"
	"meals/routes"

	"github.com/gin-gonic/gin"
)

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
