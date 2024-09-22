package routes

import (
	"meals/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// Home
	router.GET("/", controllers.HomeHandler)

	// Auth
	router.GET("/auth/:provider", controllers.AuthHandler)
	router.GET("/auth/:provider/callback", controllers.AuthCallback)
	router.GET("/auth/logout/:provider", controllers.AuthLogout)

	// Meals
	router.GET("/meals", controllers.GetMeals)
	router.GET("/meals/:id", controllers.GetMealByID)
	router.POST("/meals", controllers.PostMeals)
	router.DELETE("/meals/:id", controllers.DeleteMealByID)

	// Menus
	router.POST("/menus", controllers.PostMenu)
	router.PUT("/menus", controllers.UpdateMenu)
}
