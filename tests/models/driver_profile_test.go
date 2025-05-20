package models_test

import (
	"fmt"
	"meals/models"
	"meals/tests"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDriverProfileManagement(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("CreateDriverProfile", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver user
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a profile for the driver
		sedan := "sedan"
		license := "DL12345678"
		isAvailable := true
		profile := models.UserProfile{
			UserID:         fmt.Sprintf("%d", driver.ID),
			PhoneNumber:    "555-123-4567",
			VehicleType:    &sedan,
			LicenseNumber:  &license,
			IsAvailable:    &isAvailable,
		}
		
		// Save the profile
		err := db.Create(&profile).Error
		assert.Nil(t, err, "Expected no error when creating driver profile")
		
		// Retrieve the profile
		var retrievedProfile models.UserProfile
		err = db.Where("user_id = ?", driver.ID).First(&retrievedProfile).Error
		assert.Nil(t, err, "Expected no error when retrieving driver profile")
		assert.Equal(t, profile.VehicleType, retrievedProfile.VehicleType)
		assert.Equal(t, profile.LicenseNumber, retrievedProfile.LicenseNumber)
		assert.True(t, *retrievedProfile.IsAvailable)
	})
	
	t.Run("UpdateDriverAvailability", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver user
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a profile for the driver
		suv := "suv"
		license := "DL98765432"
		isAvailable := true
		profile := models.UserProfile{
			UserID:         fmt.Sprintf("%d", driver.ID),
			PhoneNumber:    "555-123-4567",
			VehicleType:    &suv,
			LicenseNumber:  &license,
			IsAvailable:    &isAvailable,
		}
		
		// Save the profile
		db.Create(&profile)
		
		// Update driver availability to unavailable
		unavailable := false
		profile.IsAvailable = &unavailable
		err := db.Save(&profile).Error
		assert.Nil(t, err, "Expected no error when updating driver availability")
		
		// Retrieve the updated profile
		var retrievedProfile models.UserProfile
		err = db.Where("user_id = ?", driver.ID).First(&retrievedProfile).Error
		assert.Nil(t, err, "Expected no error when retrieving updated driver profile")
		assert.False(t, *retrievedProfile.IsAvailable, "Expected driver to be unavailable")
	})
	
	t.Run("DriverDeliveryAssignment", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a driver profile
		sedan := "sedan"
		license := "DL12345678"
		isAvailable := true
		driverProfile := models.UserProfile{
			UserID:         fmt.Sprintf("%d", driver.ID),
			PhoneNumber:    "555-123-4567",
			VehicleType:    &sedan,
			LicenseNumber:  &license,
			IsAvailable:    &isAvailable,
		}
		db.Create(&driverProfile)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)
		
		// Create an order
		order := models.Order{
			UserID:         fmt.Sprintf("%d", customer.ID),
			DeliveryAddress: "123 Test St",
			Status:          models.OrderStatusPending,
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			TotalAmount:     meal.Price,                    // Set initial total amount
		}

		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 1,
			Price:    meal.Price,
			Name:     meal.Name,
		}


		// Save the order with items in a transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			// Create the order first
			if err := tx.Create(&order).Error; err != nil {
				return err
			}

			// Set order ID for the order item and create it
			orderItem.OrderID = order.ID
			if err := tx.Create(&orderItem).Error; err != nil {
				return err
			}

			// Update order status to preparing
			order.Status = models.OrderStatusPreparing
			return tx.Save(&order).Error
		})


		assert.Nil(t, err, "Expected no error when creating order with items")

		// Retrieve the updated order
		var retrievedOrder models.Order
		err = db.Preload("OrderItems").First(&retrievedOrder, order.ID).Error
		assert.Nil(t, err, "Expected no error when retrieving updated order")
		assert.Equal(t, models.OrderStatusPreparing, retrievedOrder.Status, "Expected order status to be updated to preparing")
		assert.Len(t, retrievedOrder.OrderItems, 1, "Expected order to have one item")
	})
	
	t.Run("DriverOrderAssignment", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create two drivers
		driver1 := tests.CreateTestUser(db, models.UserTypeDriver)
		driver2 := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create driver profiles
		sedan1 := "sedan"
		license1 := "DL11111111"
		isAvailable1 := true
		profile1 := models.UserProfile{
			UserID:         fmt.Sprintf("%d", driver1.ID),
			PhoneNumber:    "555-111-1111",
			VehicleType:    &sedan1,
			LicenseNumber:  &license1,
			IsAvailable:    &isAvailable1,
		}
		db.Create(&profile1)
		
		suv2 := "suv"
		license2 := "DL22222222"
		isAvailable2 := true
		profile2 := models.UserProfile{
			UserID:         fmt.Sprintf("%d", driver2.ID),
			PhoneNumber:    "555-222-2222",
			VehicleType:    &suv2,
			LicenseNumber:  &license2,
			IsAvailable:    &isAvailable2,
		}
		db.Create(&profile2)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)
		
		// Create two orders
		orderItem1 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 1,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		order1 := models.Order{
			UserID:         fmt.Sprintf("%d", customer.ID),
			Status:          models.OrderStatusPaid,
			DeliveryAddress: "123 East St, Test City, Test State 12345",
			DeliveryDate:    time.Now().Add(24 * time.Hour),
		}
		db.Create(&order1)
		orderItem1.OrderID = order1.ID
		db.Create(&orderItem1)
		
		orderItem2 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		order2 := models.Order{
			UserID:         fmt.Sprintf("%d", customer.ID),
			Status:          models.OrderStatusPaid,
			DeliveryAddress: "456 West St, Test City, Test State 67890",
			DeliveryDate:    time.Now().Add(48 * time.Hour),
		}
		db.Create(&order2)
		orderItem2.OrderID = order2.ID
		db.Create(&orderItem2)
		
		// Update orders with driver information in their status or notes
		// Since DriverID is no longer a field, we'll update the status to indicate it's being prepared
		order1.Status = models.OrderStatusPreparing
		order1.DeliveryNotes = "Assigned to driver1"
		db.Save(&order1)
		
		order2.Status = models.OrderStatusPreparing
		order2.DeliveryNotes = "Assigned to driver2"
		db.Save(&order2)
		
		// Since we no longer have a direct driver-order relationship,
		// we'll verify the orders are in the preparing state
		var preparingOrders []models.Order
		db.Where("status = ?", models.OrderStatusPreparing).Find(&preparingOrders)
		assert.Len(t, preparingOrders, 2, "Expected 2 orders to be in preparing state")
		
		// Verify the orders have the correct notes
		var updatedOrder1, updatedOrder2 models.Order
		db.First(&updatedOrder1, order1.ID)
		db.First(&updatedOrder2, order2.ID)
		
		assert.Equal(t, "Assigned to driver1", updatedOrder1.DeliveryNotes, "Expected order1 to have driver1 note")
		assert.Equal(t, "Assigned to driver2", updatedOrder2.DeliveryNotes, "Expected order2 to have driver2 note")
		assert.Equal(t, models.OrderStatusPreparing, updatedOrder1.Status, "Expected order1 to be in preparing status")
		assert.Equal(t, models.OrderStatusPreparing, updatedOrder2.Status, "Expected order2 to be in preparing status")
	})
}
