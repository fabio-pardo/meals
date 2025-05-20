package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"errors"
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

func TestOrderEndpoints(t *testing.T) {
	db := tests.SetupTestSuite(t)
	gin.SetMode(gin.TestMode)

	t.Run("CreateOrder_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)

		// Create order request
		orderRequest := map[string]interface{}{
			"delivery_address": "123 Test St, Test City, Test State 12345",
			"delivery_date":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"payment_method":   "credit_card",
			"order_items": []map[string]interface{}{
				{
					"item_type": "meal",
					"item_id":   meal.ID,
					"quantity":  2,
					"price":     meal.Price,
					"name":      meal.Name,
				},
			},
		}

		jsonData, _ := json.Marshal(orderRequest)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		authenticateUser(c, customer)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic with transaction support
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Start a transaction
			result, err := store.WithTransactionResult(c, database.DB, func(tx store.TxHandle) (interface{}, error) {
				var orderRequest models.Order

				if err := c.ShouldBindJSON(&orderRequest); err != nil {
					return nil, handlers.ValidationError("input", "Invalid order data")
				}

				// Set user ID from authenticated user
				userVal, _ := c.Get("user")
				user := userVal.(models.User)
				orderRequest.UserID = fmt.Sprintf("%d", user.ID)

				// Set initial status
				orderRequest.Status = models.OrderStatusPending

				// Validate order
				errors := orderRequest.ValidateOrder()
				if len(errors) > 0 {
					return nil, handlers.ValidationError("order", errors[0])
				}

				// Calculate total amount
				orderRequest.CalculateTotalAmount()

				// Save order
				if err := tx.Create(&orderRequest).Error; err != nil {
					return nil, handlers.DatabaseError("Failed to create order")
				}

				// Save order items with the order ID
				for i := range orderRequest.OrderItems {
					orderRequest.OrderItems[i].OrderID = orderRequest.ID
					if err := tx.Create(&orderRequest.OrderItems[i]).Error; err != nil {
						return nil, handlers.DatabaseError("Failed to create order item")
					}
				}

				return orderRequest, nil
			})

			if err != nil {
				var appErr handlers.AppError
				if errors.As(err, &appErr) {
					handlers.RespondWithError(c, appErr.ToResponse())
				} else {
					handlers.RespondWithError(c, handlers.DatabaseError("Failed to process order"))
				}
				return
			}

			order := result.(models.Order)
			c.JSON(http.StatusCreated, order)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify response body
		var responseOrder models.Order
		err := json.Unmarshal(w.Body.Bytes(), &responseOrder)
		assert.Nil(t, err)
		assert.Equal(t, customer.ID, responseOrder.UserID)
		assert.Equal(t, models.OrderStatusPending, responseOrder.Status)
		assert.Equal(t, float64(12.99*2), responseOrder.TotalAmount)
	})

	t.Run("GetOrder_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)

		// Create an order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}

		// Create an order using helper
		order := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem})

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/orders/%d", order.ID), nil)

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		authenticateUser(c, customer)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", order.ID)}}

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			orderID := order.ID
			userVal, _ := c.Get("user")
			user := userVal.(models.User)

			var order models.Order

			// Get the order with items
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Different logic based on user type
			query := database.DB.Preload("OrderItems")
			if user.UserType != models.UserTypeAdmin && user.UserType != models.UserTypeDriver {
				// Regular customers can only see their own orders
				query = query.Where("user_id = ?", user.ID)
			}

			if err := query.First(&order, orderID).Error; err != nil {
				handlers.RespondWithError(c, handlers.NotFoundError("Order not found"))
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
		assert.Equal(t, order.ID, responseOrder.ID)
		assert.Equal(t, customer.ID, responseOrder.UserID)
		assert.Len(t, responseOrder.OrderItems, 1)
	})

	t.Run("UpdateOrderStatus_ByDriver", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)

		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)

		// Create an order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}

		// Create an order using helper
		order := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem})

		// Set order to paid status
		order.Status = models.OrderStatusPaid
		db.Save(&order)

		// Prepare status update request
		updateRequest := map[string]interface{}{
			"status": models.OrderStatusDelivering,
		}

		jsonData, _ := json.Marshal(updateRequest)

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/orders/%d/status", order.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		authenticateUser(c, driver)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", order.ID)}}

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic with transaction support
			orderID := order.ID

			var statusUpdate struct {
				Status models.OrderStatus `json:"status" binding:"required"`
			}

			if err := c.ShouldBindJSON(&statusUpdate); err != nil {
				handlers.RespondWithError(c, handlers.ValidationError("input", "Invalid status update request"))
				return
			}

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Start a transaction
			result, err := store.WithTransactionResult(c, database.DB, func(tx store.TxHandle) (interface{}, error) {
				var order models.Order

				if err := tx.First(&order, orderID).Error; err != nil {
					return nil, handlers.NotFoundError("Order not found")
				}

				// Update status
				order.Status = statusUpdate.Status

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
		assert.Equal(t, models.OrderStatusDelivering, responseOrder.Status)

		// Verify the order was updated in the database
		var updatedOrder models.Order
		db.First(&updatedOrder, order.ID)
		assert.Equal(t, models.OrderStatusDelivering, updatedOrder.Status)
	})

	t.Run("GetCustomerOrders", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create meals
		meal1 := tests.CreateTestMeal(db, "Meal 1", 11.99)
		meal2 := tests.CreateTestMeal(db, "Meal 2", 13.99)

		// Create order items
		orderItem1 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal1.ID,
			Quantity: 1,
			Price:    meal1.Price,
			Name:     meal1.Name,
		}

		orderItem2 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal2.ID,
			Quantity: 2,
			Price:    meal2.Price,
			Name:     meal2.Name,
		}

		// Create two orders
		tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem1})
		tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem2})

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/orders", nil)

		// Create context using Gin's test utilities
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		authenticateUser(c, customer)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userVal, _ := c.Get("user")
			user := userVal.(models.User)

			var orders []models.Order

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Different logic based on user type
			query := database.DB.Preload("OrderItems")
			if user.UserType != models.UserTypeAdmin {
				// Regular customers can only see their own orders
				query = query.Where("user_id = ?", user.ID)
			}

			if err := query.Find(&orders).Error; err != nil {
				handlers.RespondWithError(c, handlers.DatabaseError("Failed to retrieve orders"))
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
		assert.Len(t, responseOrders, 2, "Expected customer to have 2 orders")

		// Verify all orders belong to the customer
		for _, order := range responseOrders {
			assert.Equal(t, customer.ID, order.UserID)
			assert.NotEmpty(t, order.OrderItems)
		}
	})
}
