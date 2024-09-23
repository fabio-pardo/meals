package main

import (
	"meals/auth"
	"meals/config"
	"meals/routes"
	"meals/store"
)

func main() {
	// Init config
	config.InitConfig()

	// Init OAuth2
	auth.InitOAuth2()

	// Initialize the database
	store.InitDB()

	// Init App
	routes.InitRouter()
}
