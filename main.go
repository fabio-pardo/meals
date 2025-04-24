package main

import (
	"log"
	"meals/auth"
	"meals/config"
	"meals/routes"
	"meals/store"
)

func main() {
	// Initialize configuration first
	log.Println("Initializing configuration...")
	config.InitConfig()

	// Initialize the DBs (Postgres and Redis)
	log.Println("Initializing databases...")
	store.InitStores()

	// Initialize OAuth2 with the loaded configuration
	log.Println("Initializing OAuth2...")
	auth.InitOAuth2()

	// Initialize and start the router
	log.Println("Starting web server...")
	routes.InitRouter()
}
