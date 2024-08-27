package routes

import (
	"meals/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/meals", controllers.GetMeals)
	router.GET("/meals/:id", controllers.GetMealByID)
	router.POST("/meals", controllers.PostMeals)
}
