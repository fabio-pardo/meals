package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Meal struct {
	CreatedAt time.Time `gorm:"type:timestamp; default:current_timestamp"`
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name" gorm:"size:255; not null"`
	Price     float64   `json:"price" gorm:"not null"`
}

var meals = []Meal{
	{ID: "1", Name: "Lasagna", Price: 10.99},
	{ID: "2", Name: "Pizza", Price: 10.99},
	{ID: "3", Name: "Sushi", Price: 10.99},
}

func main() {
	migrateDB()

	router := gin.Default()
	router.GET("/meals", getMeals)
	router.GET("/meals/:id", getMealById)
	router.POST("/meals", postMeals)

	router.Run("localhost:8080")
}

func migrateDB() *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the DB")
	}

	if err := db.AutoMigrate(&Meal{}); err != nil {
		panic("Failed to migrate database schema")
	}
	return db
}

// getMeals responds with the list of all meals as JSON.
func getMeals(c *gin.Context) {
	c.JSON(http.StatusOK, meals)
}

func postMeals(c *gin.Context) {
	var newMeal Meal

	// Call BindJSON to bind the received JSON to newMeal
	if err := c.BindJSON(&newMeal); err != nil {
		return
	}

	meals = append(meals, newMeal)
	c.JSON(http.StatusCreated, newMeal)
}

func getMealById(c *gin.Context) {
	id := c.Param("id")

	for _, m := range meals {
		if m.ID == id {
			c.IndentedJSON(http.StatusOK, m)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "meal not found"})
}
