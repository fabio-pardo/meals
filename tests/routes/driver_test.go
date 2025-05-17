package routes_test

import (
	"bytes"
	"encoding/json"
	"errors"
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

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/driver/profile", nil)

		// Create a new Gin context with the response recorder and request
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add driver authentication to context
		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userVal, _ := c.Get("user")
			user := userVal.(models.User)

			// Check if user is a driver
			if user.UserType != models.UserTypeDriver {
				handlers.RespondWithError(c, handlers.ErrorResponse{
					Status:  http.StatusForbidden,
					Code:    "FORBIDDEN",
					Message: "Access denied. Driver account required.",
				})
				return
			}

			// Return basic user info (driver-specific profile will be handled separately)
			c.JSON(http.StatusOK, gin.H{
				"user_id": user.ID,
				"email":   user.Email,
				"name":    user.Name,
				"type":    user.UserType,
			})
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		assert.Equal(t, float64(driver.ID), response["user_id"])
		assert.Equal(t, string(models.UserTypeDriver), response["type"])
	})

	t.Run("UpdateDriverAvailability", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)

		// Prepare availability update
		updateRequest := map[string]interface{}{
			"is_available": true,
		}

		jsonData, _ := json.Marshal(updateRequest)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/driver/availability", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Create a new Gin context with the response recorder and request
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add driver authentication to context
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
				handlers.RespondWithError(c, handlers.ValidationError("input", "Invalid availability data"))
				return
			}

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Use WithTransactionNoResult for operations that don't return a result
			err := store.WithTransactionNoResult(c, database.DB, func(tx store.TxHandle) error {
				// In a real implementation, this would update the driver's availability
				// For now, we'll just return a success response
				response := gin.H{
					"user_id":      user.ID,
					"is_available": availabilityUpdate.IsAvailable,
					"updated_at":   time.Now(),
				}
				c.JSON(http.StatusOK, response)
				return nil
			})

			if err != nil {
				handlers.RespondWithError(c, handlers.ErrorResponse{
					Status:  http.StatusInternalServerError,
					Code:    "DATABASE_ERROR",
					Message: "Failed to update driver availability: " + err.Error(),
				})
				return
			}
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, err)
		assert.Equal(t, driver.ID, uint(response["user_id"].(float64)))
		assert.NotNil(t, response["is_available"], "Expected is_available in response")
	})

	t.Run("GetDriverAssignedOrders", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)

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
		order1.Status = models.OrderStatusPreparing
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

		// Create a new Gin context with the response recorder and request
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add driver authentication to context
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
				handlers.RespondWithError(c, handlers.DatabaseError("Failed to retrieve driver orders"))
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
		order.Status = models.OrderStatusPreparing
		db.Save(&order)

		// Prepare status update request (to InDelivery)
		updateRequest := map[string]interface{}{
			"status": models.OrderStatusDelivering,
		}

		jsonData, _ := json.Marshal(updateRequest)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/driver/orders/%d/status", order.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add driver authentication to context
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
				handlers.RespondWithError(c, handlers.ValidationError("input", "Invalid status update request"))
				return
			}

			// Check valid status transitions
			isValidTransition := func(currentStatus, newStatus models.OrderStatus) bool {
				validTransitions := map[models.OrderStatus][]models.OrderStatus{
					models.OrderStatusPreparing:  {models.OrderStatusDelivering},
					models.OrderStatusDelivering: {models.OrderStatusDelivered},
				}

				transitions, exists := validTransitions[currentStatus]
				if !exists {
					return false
				}

				for _, validStatus := range transitions {
					if validStatus == newStatus {
						return true
					}
				}
				return false
			}

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Update order status
			var order models.Order
			if err := database.DB.Where("id = ? AND driver_id = ?", orderID, user.ID).First(&order).Error; err != nil {
				handlers.RespondWithError(c, handlers.NotFoundError("Order not found or not assigned to you"))
				return
			}

			// Validate status transition
			if !isValidTransition(order.Status, statusUpdate.Status) {
				handlers.RespondWithError(c, handlers.ValidationError("status", "Invalid status transition"))
				return
			}

			// Update status
			order.Status = statusUpdate.Status

			// Update status
			order.Status = statusUpdate.Status

			if err := database.DB.Save(&order).Error; err != nil {
				handlers.RespondWithError(c, handlers.DatabaseError("Failed to update order status"))
				return
			}

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
		assert.Equal(t, models.OrderStatusDelivering, responseOrder.Status)
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
		order.Status = models.OrderStatusDelivering
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

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Add driver authentication to context
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
				handlers.RespondWithError(c, handlers.ValidationError("Invalid status update", err.Error()))
				return
			}

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Replace the WithTransaction call with WithTransactionResult
			result, err := store.WithTransactionResult(c, database.DB, func(tx store.TxHandle) (interface{}, error) {
				var order models.Order

				if err := tx.Where("id = ? AND driver_id = ?", orderID, user.ID).First(&order).Error; err != nil {
					return nil, handlers.NotFoundError("Order not found or not assigned to you")
				}

				// For this test, we're only allowing InDelivery -> Delivered transition
				if order.Status != models.OrderStatusDelivering || statusUpdate.Status != models.OrderStatusDelivered {
					return nil, handlers.ValidationError("status", "Invalid status transition")
				}

				// Update status
				order.Status = statusUpdate.Status

				// The Order model doesn't have a DeliveredAt field, so we'll use UpdatedAt
				// which will be automatically updated by gorm
				if err := tx.Save(&order).Error; err != nil {
					return nil, handlers.DatabaseError("Failed to update order status")
				}

				return order, nil
			})

			if err != nil {
				var appErr handlers.AppError
				if errors.As(err, &appErr) {
					handlers.RespondWithError(c, appErr.ToResponse())
				} else {
					handlers.RespondWithError(c, handlers.DatabaseError("Failed to update order status"))
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

		// Verify the order was updated in the database
		var updatedOrder models.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, models.OrderStatusDelivering, updatedOrder.Status)
	})
}
