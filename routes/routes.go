package routes

import (
	"meals/auth"
	"meals/config"
	"meals/handlers"
	"meals/middleware"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// Home - handle both authenticated and non-authenticated users
	router.GET("/", func(c *gin.Context) {
		if auth.IsAuthenticated(c) {
			handlers.HomeHandler(c)
		} else {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Welcome to Meals App</title>
				<style>
					body { 
						font-family: Arial, sans-serif; 
						margin: 40px;
						line-height: 1.6;
						color: #333;
						background-color: #f5f5f5;
					}
					h1 { 
						color: #2c3e50; 
						margin-bottom: 20px;
					}
					a { 
						color: #3498db; 
						text-decoration: none;
						font-weight: bold;
					}
					a:hover { text-decoration: underline; }
					.container {
						max-width: 800px;
						margin: 0 auto;
						padding: 30px;
						border-radius: 8px;
						background-color: #fff;
						box-shadow: 0 2px 10px rgba(0,0,0,0.1);
					}
					.login-btn {
						display: inline-block;
						background-color: #3498db;
						color: white;
						padding: 10px 20px;
						border-radius: 4px;
						margin-top: 10px;
					}
					.login-btn:hover {
						background-color: #2980b9;
						text-decoration: none;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>Welcome to Meals App</h1>
					<p>Your personal meal planning assistant. Organize your weekly meals with ease and never wonder what's for dinner again.</p>
					<p>Features:</p>
					<ul>
						<li>Create and manage your meal list</li>
						<li>Plan weekly menus</li>
						<li>Organize your favorite recipes</li>
						<li>Place orders for meal delivery</li>
					</ul>
					<p>To get started, please sign in:</p>
					<a href="/auth/google" class="login-btn">Login with Google</a>
				</div>
			</body>
			</html>
			`)
		}
	})

	// Protected routes - require authentication
	router.GET("/dashboard", auth.RequireAuth(handlers.HomeHandler))

	// Auth
	router.GET("/auth/:provider", handlers.GetAuthProviderHandler)
	router.GET("/auth/:provider/callback", handlers.GetAuthCallbackHandler)
	router.GET("/logout", handlers.LogoutHandler)

	// Meals
	router.GET("/meals", handlers.GetMealsHandler)
	router.POST("/meals", handlers.CreateMealHandler)
	router.GET("/meals/:id", handlers.GetMealHandler)
	router.PUT("/meals/:id", handlers.UpdateMealHandler)
	router.DELETE("/meals/:id", handlers.DeleteMealHandler)

	// Menus
	router.GET("/menus", handlers.GetMenusHandler)
	router.POST("/menus", handlers.CreateMenuHandler)
	router.PUT("/menus", handlers.UpdateMenuHandler)

	// Orders - all routes protected with authentication
	ordersGroup := router.Group("/orders")
	ordersGroup.Use(auth.RequireAuth())
	{
		ordersGroup.POST("", handlers.CreateOrderHandler)
		ordersGroup.GET("", handlers.ListOrdersHandler)
		ordersGroup.GET("/:id", handlers.GetOrderHandler)
		ordersGroup.PUT("/:id/status", handlers.UpdateOrderStatusHandler)
		ordersGroup.POST("/:id/cancel", handlers.CancelOrderHandler)
	}
}

func InitRouter() {
	// Initialize the router without default middleware
	router := gin.New()

	// Use our custom middlewares
	router.Use(middleware.RequestID()) // Add request ID to each request
	router.Use(middleware.Logger())    // Log requests with request IDs
	router.Use(middleware.Recovery())  // Recover from panics with request ID tracking

	RegisterRoutes(router)

	// Use the server address from config
	serverAddress := config.AppConfig.Server.GetServerAddress()
	router.Run(serverAddress)
}
