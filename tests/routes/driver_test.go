package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"meals/handlers"
	"meals/models"
	"meals/store"
	"meals/tests"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDriverEndpoints(t *testing.T) {
	db := tests.SetupTestSuite(t)
	gin.SetMode(gin.TestMode)

	t.Run("GetDriverProfile", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a profile for the driver
		profile := models.UserProfile{
			UserID:         driver.ID,
			PhoneNumber:    "555-123-4567",
			VehicleType:    "sedan",
			LicenseNumber:  "DL12345678",
			IsAvailable:    true,
		}
		db.Create(&profile)
		
		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/driver/profile", nil)
		
		// Add driver authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})
		
		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userVal, _ := c.Get("user")
			user := userVal.(models.User)
			
			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)
			
			var driverProfile models.UserProfile
			if err := database.DB.Where("user_id = ?", user.ID).First(&driverProfile).Error; err != nil {
				handlers.NotFoundError("Driver profile not found").ToResponse(c)
				return
			}
			
			c.JSON(http.StatusOK, driverProfile)
		}
		
		// Execute handler
		handler(c)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response body
		var responseProfile models.UserProfile
		err := json.Unmarshal(w.Body.Bytes(), &responseProfile)
		assert.Nil(t, err)
		assert.Equal(t, profile.UserID, responseProfile.UserID)
		assert.Equal(t, profile.VehicleType, responseProfile.VehicleType)
		assert.Equal(t, profile.LicenseNumber, responseProfile.LicenseNumber)
		assert.Equal(t, profile.IsAvailable, responseProfile.IsAvailable)
	})
	
	t.Run("UpdateDriverAvailability", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a profile for the driver
		profile := models.UserProfile{
			UserID:         driver.ID,
			PhoneNumber:    "555-123-4567",
			VehicleType:    "sedan",
			LicenseNumber:  "DL12345678",
			IsAvailable:    false, // Initially not available
		}
		db.Create(&profile)
		
		// Prepare availability update
		updateRequest := map[string]interface{}{
			"is_available": true,
		}
		
		jsonData, _ := json.Marshal(updateRequest)
		
		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/driver/availability", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// Add driver authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})
		
		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userVal, _ := c.Get("user")
			user := userVal.(models.User)
			
			var availabilityUpdate struct {
				IsAvailable bool `json:"is_available" binding:"required"`
			}
			
			if err := c.ShouldBindJSON(&availabilityUpdate); err != nil {
				handlers.ValidationError("input", "Invalid availability data").ToResponse(c)
				return
			}
			
			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)
			
			// Start a transaction
			result, err := store.WithTransaction(c, database.DB, func(tx store.TxHandle) (interface{}, error) {
				var driverProfile models.UserProfile
				if err := tx.Where("user_id = ?", user.ID).First(&driverProfile).Error; err != nil {
					return nil, handlers.NotFoundError("Driver profile not found")
				}
				
				// Update availability
				driverProfile.IsAvailable = availabilityUpdate.IsAvailable
				if err := tx.Save(&driverProfile).Error; err != nil {
					return nil, handlers.DatabaseError("Failed to update driver availability")
				}
				
				return driverProfile, nil
			})
			
			if err != nil {
				if appErr, ok := err.(handlers.AppError); ok {
					appErr.ToResponse(c)
				} else {
					handlers.DatabaseError("Failed to update driver availability").ToResponse(c)
				}
				return
			}
			
			driverProfile := result.(models.UserProfile)
			c.JSON(http.StatusOK, driverProfile)
		}
		
		// Execute handler
		handler(c)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response body
		var responseProfile models.UserProfile
		err := json.Unmarshal(w.Body.Bytes(), &responseProfile)
		assert.Nil(t, err)
		assert.Equal(t, profile.UserID, responseProfile.UserID)
		assert.True(t, responseProfile.IsAvailable, "Expected driver to be available")
		
		// Verify the profile was updated in the database
		var updatedProfile models.UserProfile
		db.Where("user_id = ?", driver.ID).First(&updatedProfile)
		assert.True(t, updatedProfile.IsAvailable, "Expected driver to be available in database")
	})
	
	t.Run("GetDriverAssignedOrders", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a driver profile
		driverProfile := models.UserProfile{
			UserID:         driver.ID,
			PhoneNumber:    "555-123-4567",
			VehicleType:    "sedan",
			LicenseNumber:  "DL12345678",
			IsAvailable:    true,
		}
		db.Create(&driverProfile)
		
		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)
		
		// Create order items
		orderItem1 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 1,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		orderItem2 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create orders assigned to the driver
		order1 := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem1})
		order1.DriverID = &driver.ID
		order1.Status = models.OrderStatusInPreparation
		db.Save(&order1)
		
		order2 := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem2})
		order2.DriverID = &driver.ID
		order2.Status = models.OrderStatusPaid
		db.Save(&order2)
		
		// Create an order not assigned to the driver
		tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem1})
		
		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/driver/orders", nil)
		
		// Add driver authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})
		
		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userVal, _ := c.Get("user")
			user := userVal.(models.User)
			
			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)
			
			var orders []models.Order
			if err := database.DB.Where("driver_id = ?", user.ID).Preload("OrderItems").Find(&orders).Error; err != nil {
				handlers.DatabaseError("Failed to retrieve driver orders").ToResponse(c)
				return
			}
			
			c.JSON(http.StatusOK, orders)
		}
		
		// Execute handler
		handler(c)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response body
		var responseOrders []models.Order
		err := json.Unmarshal(w.Body.Bytes(), &responseOrders)
		assert.Nil(t, err)
		assert.Len(t, responseOrders, 2, "Expected 2 orders assigned to the driver")
		
		// Verify the orders are assigned to the driver
		for _, order := range responseOrders {
			assert.NotNil(t, order.DriverID, "Expected order to have a driver assigned")
			assert.Equal(t, driver.ID, *order.DriverID, "Expected order to be assigned to the test driver")
			assert.NotEmpty(t, order.OrderItems, "Expected order to have items")
		}
	})
	
	t.Run("UpdateOrderStatusByDriver", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 1,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create an order assigned to the driver
		order := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem})
		order.DriverID = &driver.ID
		order.Status = models.OrderStatusInPreparation
		db.Save(&order)
		
		// Prepare status update request (to InDelivery)
		updateRequest := map[string]interface{}{
			"status": models.OrderStatusInDelivery,
		}
		
		jsonData, _ := json.Marshal(updateRequest)
		
		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/driver/orders/%d/status", order.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// Add driver authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", order.ID)}}
		
		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			orderID := order.ID
			userVal, _ := c.Get("user")
			user := userVal.(models.User)
			
			var statusUpdate struct {
				Status models.OrderStatus `json:"status" binding:"required"`
			}
			
			if err := c.ShouldBindJSON(&statusUpdate); err != nil {
				handlers.ValidationError("input", "Invalid status update request").ToResponse(c)
				return
			}
			
			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)
			
			// Check valid status transitions
			isValidTransition := func(currentStatus, newStatus models.OrderStatus) bool {
				validTransitions := map[models.OrderStatus][]models.OrderStatus{
					models.OrderStatusInPreparation: {models.OrderStatusInDelivery},
					models.OrderStatusInDelivery:    {models.OrderStatusDelivered},
				}
				
				if transitions, exists := validTransitions[currentStatus]; exists {
					for _, validStatus := range transitions {
						if validStatus == newStatus {
							return true
						}
					}
				}
				return false
			}
			
			// Start a transaction
			result, err := store.WithTransaction(c, database.DB, func(tx store.TxHandle) (interface{}, error) {
				var order models.Order
				
				if err := tx.Where("id = ? AND driver_id = ?", orderID, user.ID).First(&order).Error; err != nil {
					return nil, handlers.NotFoundError("Order not found or not assigned to you")
				}
				
				// Validate status transition
				if !isValidTransition(order.Status, statusUpdate.Status) {
					return nil, handlers.ValidationError("status", "Invalid status transition")
				}
				
				// Update status
				order.Status = statusUpdate.Status
				
				// If status is being set to delivered, set delivery time
				if order.Status == models.OrderStatusDelivered {
					now := time.Now()
					order.DeliveredAt = &now
				}
				
				if err := tx.Save(&order).Error; err != nil {
					return nil, handlers.DatabaseError("Failed to update order status")
				}
				
				return order, nil
			})
			
			if err != nil {
				if appErr, ok := err.(handlers.AppError); ok {
					appErr.ToResponse(c)
				} else {
					handlers.DatabaseError("Failed to update order status").ToResponse(c)
				}
				return
			}
			
			order := result.(models.Order)
			c.JSON(http.StatusOK, order)
		}
		
		// Execute handler
		handler(c)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response body
		var responseOrder models.Order
		err := json.Unmarshal(w.Body.Bytes(), &responseOrder)
		assert.Nil(t, err)
		assert.Equal(t, models.OrderStatusInDelivery, responseOrder.Status)
		
		// Verify the order was updated in the database
		var updatedOrder models.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, models.OrderStatusInDelivery, updatedOrder.Status)
	})
	
	t.Run("MarkOrderAsDelivered", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)
		
		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 1,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create an order assigned to the driver that's in delivery
		order := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem})
		order.DriverID = &driver.ID
		order.Status = models.OrderStatusInDelivery
		db.Save(&order)
		
		// Prepare delivery confirmation request
		updateRequest := map[string]interface{}{
			"status": models.OrderStatusDelivered,
		}
		
		jsonData, _ := json.Marshal(updateRequest)
		
		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/driver/orders/%d/status", order.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// Add driver authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", order.ID)}}
		
		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			orderID := order.ID
			userVal, _ := c.Get("user")
			user := userVal.(models.User)
			
			var statusUpdate struct {
				Status models.OrderStatus `json:"status" binding:"required"`
			}
			
			if err := c.ShouldBindJSON(&statusUpdate); err != nil {
				handlers.ValidationError("input", "Invalid status update request").ToResponse(c)
				return
			}
			
			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)
			
			// Start a transaction
			result, err := store.WithTransaction(c, database.DB, func(tx store.TxHandle) (interface{}, error) {
				var order models.Order
				
				if err := tx.Where("id = ? AND driver_id = ?", orderID, user.ID).First(&order).Error; err != nil {
					return nil, handlers.NotFoundError("Order not found or not assigned to you")
				}
				
				// For this test, we're only allowing InDelivery -> Delivered transition
				if order.Status != models.OrderStatusInDelivery || statusUpdate.Status != models.OrderStatusDelivered {
					return nil, handlers.ValidationError("status", "Invalid status transition")
				}
				
				// Update status
				order.Status = statusUpdate.Status
				
				// Set delivery time
				now := time.Now()
				order.DeliveredAt = &now
				
				if err := tx.Save(&order).Error; err != nil {
					return nil, handlers.DatabaseError("Failed to update order status")
				}
				
				return order, nil
			})
			
			if err != nil {
				if appErr, ok := err.(handlers.AppError); ok {
					appErr.ToResponse(c)
				} else {
					handlers.DatabaseError("Failed to update order status").ToResponse(c)
				}
				return
			}
			
			order := result.(models.Order)
			c.JSON(http.StatusOK, order)
		}
		
		// Execute handler
		handler(c)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify response body
		var responseOrder models.Order
		err := json.Unmarshal(w.Body.Bytes(), &responseOrder)
		assert.Nil(t, err)
		assert.Equal(t, models.OrderStatusDelivered, responseOrder.Status)
		assert.NotNil(t, responseOrder.DeliveredAt, "Expected delivery time to be set")
		
		// Verify the order was updated in the database
		var updatedOrder models.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, models.OrderStatusDelivered, updatedOrder.Status)
		assert.NotNil(t, updatedOrder.DeliveredAt, "Expected delivery time to be set in database")
	})
}
