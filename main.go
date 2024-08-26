package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Meal struct {
	CreatedAt time.Time `gorm:"type:timestamp; default:current_timestamp"`
	Name      string    `json:"name" gorm:"size:255; not null"`
	ID        uint      `gorm:"primaryKey; autoIncrement; not null"`
	Price     float64   `json:"price" gorm:"not null"`
}

func main() {
	initDB()

	router := gin.Default()
	router.GET("/meals", getMeals)
	router.GET("/meals/:id", getMealById)
	router.POST("/meals", postMeals)

	router.Run("localhost:8080")
}

func initDB() {
	dsn := "host=localhost user=test password=test dbname=test port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the DB: " + err.Error())
	}
	if err := DB.AutoMigrate(&Meal{}); err != nil {
		panic("Failed to migrate database schema: " + err.Error())
	}
}

func getMeals(c *gin.Context) {
	var meals []Meal
	if err := DB.Find(&meals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meals"})
		return
	}
	c.JSON(http.StatusOK, meals)
}

func postMeals(c *gin.Context) {
	var newMeal Meal

	if err := c.BindJSON(&newMeal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := DB.Create(&newMeal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meal"})
		return
	}

	c.JSON(http.StatusCreated, newMeal)
}

func getMealById(c *gin.Context) {
	id := c.Param("id")
	var meal Meal

	if err := DB.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "meal not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meal"})
		}
		return
	}

	c.JSON(http.StatusOK, meal)
}
