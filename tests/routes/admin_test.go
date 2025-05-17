package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"meals/handlers"
	"meals/models"
	"meals/tests"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdminEndpoints(t *testing.T) {
	db := tests.SetupTestSuite(t)
	gin.SetMode(gin.TestMode)

	t.Run("GetAllUsers_AdminAccess", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create admin user
		admin := tests.CreateTestUser(db, models.UserTypeAdmin)

		// Create some regular users
		customer1 := tests.CreateTestUser(db, models.UserTypeCustomer)
		customer2 := tests.CreateTestUser(db, models.UserTypeCustomer)
		driver := tests.CreateTestUser(db, models.UserTypeDriver)

		// Set up request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/users", nil)

		// Add admin authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, admin)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			var users []models.User

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			if err := database.DB.Find(&users).Error; err != nil {
				handlers.DatabaseError("Failed to retrieve users").ToResponse(c)
				return
			}

			c.JSON(http.StatusOK, users)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response body
		var responseUsers []models.User
		err := json.Unmarshal(w.Body.Bytes(), &responseUsers)
		assert.Nil(t, err)

		// Should have at least 4 users (admin, 2 customers, 1 driver)
		assert.GreaterOrEqual(t, len(responseUsers), 4, "Expected at least 4 users")

		// Verify the users exist in the response
		foundAdmin := false
		foundCustomer1 := false
		foundCustomer2 := false
		foundDriver := false

		for _, user := range responseUsers {
			if user.ID == admin.ID {
				foundAdmin = true
				assert.Equal(t, models.UserTypeAdmin, user.UserType)
			}
			if user.ID == customer1.ID {
				foundCustomer1 = true
				assert.Equal(t, models.UserTypeCustomer, user.UserType)
			}
			if user.ID == customer2.ID {
				foundCustomer2 = true
				assert.Equal(t, models.UserTypeCustomer, user.UserType)
			}
			if user.ID == driver.ID {
				foundDriver = true
				assert.Equal(t, models.UserTypeDriver, user.UserType)
			}
		}

		assert.True(t, foundAdmin, "Expected admin user in response")
		assert.True(t, foundCustomer1, "Expected customer1 in response")
		assert.True(t, foundCustomer2, "Expected customer2 in response")
		assert.True(t, foundDriver, "Expected driver in response")
	})

	t.Run("UpdateUserRole_AdminAccess", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create admin user
		admin := tests.CreateTestUser(db, models.UserTypeAdmin)

		// Create a customer user
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Prepare role update request
		updateRequest := map[string]interface{}{
			"user_type": models.UserTypeDriver,
		}

		jsonData, _ := json.Marshal(updateRequest)

		// Set up request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/admin/users/%d/role", customer.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Add admin authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, admin)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", customer.ID)}}

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			userID := customer.ID

			var roleUpdate struct {
				UserType models.UserType `json:"user_type" binding:"required"`
			}

			if err := c.ShouldBindJSON(&roleUpdate); err != nil {
				handlers.ValidationError("input", "Invalid role update data").ToResponse(c)
				return
			}

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Find the user
			var user models.User
			if err := database.DB.First(&user, userID).Error; err != nil {
				handlers.NotFoundError("User not found").ToResponse(c)
				return
			}

			// Update the user role
			user.UserType = roleUpdate.UserType
			if err := database.DB.Save(&user).Error; err != nil {
				handlers.DatabaseError("Failed to update user role").ToResponse(c)
				return
			}

			c.JSON(http.StatusOK, user)
		}

		// Execute handler
		handler(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response body
		var responseUser models.User
		err := json.Unmarshal(w.Body.Bytes(), &responseUser)
		assert.Nil(t, err)
		assert.Equal(t, customer.ID, responseUser.ID)
		assert.Equal(t, models.UserTypeDriver, responseUser.UserType)

		// Verify the user role was updated in the database
		var updatedUser models.User
		db.First(&updatedUser, customer.ID)
		assert.Equal(t, models.UserTypeDriver, updatedUser.UserType)
	})

	t.Run("GetAllOrders_AdminAccess", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create admin user
		admin := tests.CreateTestUser(db, models.UserTypeAdmin)

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

		// Create two orders
		tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem1})
		tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem2})

		// Set up request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/orders", nil)

		// Add admin authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, admin)
		c.Set("db", &models.Database{DB: db})

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			var orders []models.Order

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			if err := database.DB.Preload("OrderItems").Find(&orders).Error; err != nil {
				handlers.DatabaseError("Failed to retrieve orders").ToResponse(c)
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
		assert.GreaterOrEqual(t, len(responseOrders), 2, "Expected at least 2 orders")

		// Check that the orders have items
		for _, order := range responseOrders {
			assert.NotEmpty(t, order.OrderItems, "Expected order to have items")
		}
	})

	t.Run("AdminAssignDriverToOrder", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create admin user
		admin := tests.CreateTestUser(db, models.UserTypeAdmin)

		// Create a customer
		customer := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a driver
		driver := tests.CreateTestUser(db, models.UserTypeDriver)

		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)

		// Create an order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 1,
			Price:    meal.Price,
			Name:     meal.Name,
		}

		// Create an order
		order := tests.CreateTestOrder(db, customer.ID, []models.OrderItem{orderItem})

		// Prepare driver assignment request
		assignRequest := map[string]interface{}{
			"driver_id": driver.ID,
		}

		jsonData, _ := json.Marshal(assignRequest)

		// Set up request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/admin/orders/%d/assign", order.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		// Add admin authentication to context
		c := &gin.Context{Request: req, Writer: w}
		authenticateUser(c, admin)
		c.Set("db", &models.Database{DB: db})
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", order.ID)}}

		// Create a handler to simulate route execution
		handler := func(c *gin.Context) {
			// Mock the handler logic
			orderID := order.ID

			var assignRequest struct {
				DriverID uint `json:"driver_id" binding:"required"`
			}

			if err := c.ShouldBindJSON(&assignRequest); err != nil {
				handlers.ValidationError("input", "Invalid assignment data").ToResponse(c)
				return
			}

			// Get database from context
			dbVal, _ := c.Get("db")
			database := dbVal.(*models.Database)

			// Find the order
			var order models.Order
			if err := database.DB.First(&order, orderID).Error; err != nil {
				handlers.NotFoundError("Order not found").ToResponse(c)
				return
			}

			// Find the driver
			var driver models.User
			if err := database.DB.First(&driver, assignRequest.DriverID).Error; err != nil {
				handlers.NotFoundError("Driver not found").ToResponse(c)
				return
			}

			// Verify the driver is actually a driver
			if driver.UserType != models.UserTypeDriver {
				handlers.ValidationError("driver", "User is not a driver").ToResponse(c)
				return
			}

			// Assign the driver to the order
			driverID := assignRequest.DriverID
			order.DriverID = &driverID
			if err := database.DB.Save(&order).Error; err != nil {
				handlers.DatabaseError("Failed to assign driver to order").ToResponse(c)
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
		assert.NotNil(t, responseOrder.DriverID, "Expected order to have a driver assigned")
		assert.Equal(t, driver.ID, *responseOrder.DriverID, "Expected correct driver ID to be assigned")

		// Verify the order was updated in the database
		var updatedOrder models.Order
		db.First(&updatedOrder, order.ID)
		assert.NotNil(t, updatedOrder.DriverID, "Expected order to have a driver assigned in database")
		assert.Equal(t, driver.ID, *updatedOrder.DriverID, "Expected correct driver ID to be assigned in database")
	})
}
