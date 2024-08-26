package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Meal struct {
	CreatedAt time.Time `gorm:"type:timestamp; default:current_timestamp"`
	Name      string    `json:"name" gorm:"size:255; not null"`
	ID        uint      `gorm:"primaryKey; autoIncrement; not null"`
	Price     float64   `json:"price" gorm:"not null"`
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
	db, err := connectDB()
	if err != nil {
		panic("Failed to connect to the DB")
	}

	if err := db.AutoMigrate(&Meal{}); err != nil {
		panic("Failed to migrate database schema")
	}
	return db
}

func connectDB() (*gorm.DB, error) {
	dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}

func getMeals(c *gin.Context) {
	c.JSON(http.StatusOK, "")
}

func postMeals(c *gin.Context) {
	var newMeal Meal

	db, err := connectDB()
	if err != nil {
		panic("Failed to connect to DB")
	}

	// Bind the received JSON to newMeal (excluding ID)
	if err := c.BindJSON(&newMeal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := db.Create(&newMeal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meal"})
		return
	}

	c.JSON(http.StatusCreated, newMeal)
}

func getMealById(c *gin.Context) {
	id := c.Param("id")
	var meal Meal

	db, err := connectDB()
	if err != nil {
		panic("Failed to connect to DB")
	}

	if err := db.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "meal not found"})
		} else {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meal"})
		}
		return
	}

	fmt.Printf("hey reached here")
	c.IndentedJSON(http.StatusFound, meal)
}
