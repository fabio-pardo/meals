package routes

import (
	"meals/auth"
	"meals/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// Home
	router.GET("/", auth.RequireAuth(handlers.HomeHandler))

	// Auth
	router.GET("/auth/:provider", handlers.GetAuthProviderHandler)
	router.GET("/auth/:provider/callback", handlers.GetAuthCallbackHandler)

	// Meals
	router.GET("/meals", handlers.GetMealsHandler)
	router.POST("/meals", handlers.CreateMealHandler)
	router.GET("/meals/:id", handlers.GetMealHandler)
	router.PUT("/meals/:id", handlers.UpdateMealHandler)
	router.DELETE("/meals/:id", handlers.DeleteMealHandler)

	// Menus
	router.POST("/menus", handlers.CreateMenuHandler)
	router.PUT("/menus", handlers.UpdateMenuHandler)
}

func InitRouter() {
	router := gin.Default()
	RegisterRoutes(router)
	router.Run("localhost:8080")
}
