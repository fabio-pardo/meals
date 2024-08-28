package controllers

import (
	"meals/config"
	"meals/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type meal_menu struct {
	ID          uint   `json:"id"`
	DeliveryDay string `json:"delivery_day"`
}

type menu struct {
	Meals []meal_menu `json:"meals"`
}

func PostMenu(c *gin.Context) {
	var menu menu
	if err := c.BindJSON(&menu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	createMenu := func(tx *gorm.DB) error {
		var m models.Menu // TODO: Accept WeekStartDate and WeekEndDate values
		if err := config.DB.Create(&m).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
			return err
		}

		var menu_meals []models.MenuMeal
		for i := 0; i < len(menu.Meals); i++ {
			menu_meal := models.MenuMeal{
				DeliveryDay: menu.Meals[i].DeliveryDay,
				MenuID:      m.ID,
				MealID:      menu.Meals[i].ID,
			}
			menu_meals = append(menu_meals, menu_meal)
		}
		if err := config.DB.CreateInBatches(&menu_meals, len(menu_meals)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meals for menu"})
			return err
		}
		return nil
	}

	config.DB.Transaction(createMenu)
}
