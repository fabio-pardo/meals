package routes

import (
	"meals/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// Meals
	router.GET("/meals", controllers.GetMeals)
	router.GET("/meals/:id", controllers.GetMealByID)
	router.POST("/meals", controllers.PostMeals)
	router.DELETE("/meals/:id", controllers.DeleteMealByID)

	// Menus
	router.POST("/menus", controllers.PostMenu)
}
