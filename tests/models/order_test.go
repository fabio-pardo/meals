package models_test

import (
	"meals/models"
	"meals/tests"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderValidation(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("ValidOrder", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		// Ensure user was created successfully
		assert.NotZero(t, user.ID, "User ID should not be zero")
		t.Logf("Created test user with ID: %d, Email: %s", user.ID, user.Email)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 9.99)
		assert.NotZero(t, meal.ID, "Meal ID should not be zero")
		t.Logf("Created test meal with ID: %d, Name: %s, Price: %.2f", meal.ID, meal.Name, meal.Price)
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		t.Logf("Created order item: %+v", orderItem)
		
		// Create order with all required fields
		order := models.Order{
			UserID:          user.ID,
			Status:          models.OrderStatusPending,
			DeliveryAddress: "123 Test St, Test City, Test State 12345",
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			OrderItems:      []models.OrderItem{orderItem},
			PaymentMethod:   "credit_card",
		}
		t.Logf("Created order: %+v", order)
		
		// Calculate total amount
		order.CalculateTotalAmount()
		t.Logf("Calculated total amount: %.2f", order.TotalAmount)
		
		// Validate order
		errors := order.ValidateOrder()
		if len(errors) > 0 {
			t.Logf("Validation errors: %v", errors)
		}
		assert.Empty(t, errors, "Expected no validation errors")
		assert.Equal(t, 19.98, order.TotalAmount, "Expected total amount to be 19.98")
		
		// Save the order to the database to ensure it validates with the database constraints
		result := db.Create(&order)
		assert.NoError(t, result.Error, "Failed to save order to database")
		assert.NotZero(t, order.ID, "Order ID should be set after saving to database")
	})
	
	t.Run("InvalidOrder_NoUserID", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 9.99)
		assert.NotZero(t, meal.ID, "Meal ID should not be zero")
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create order with all required fields except UserID
		order := models.Order{
			// UserID is intentionally left out
			Status:          models.OrderStatusPending,
			DeliveryAddress: "123 Test St, Test City, Test State 12345",
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			OrderItems:      []models.OrderItem{orderItem},
			PaymentMethod:   "credit_card",
		}
		
		// Calculate total amount
		order.CalculateTotalAmount()
		
		// Validate order
		errors := order.ValidateOrder()
		if len(errors) == 0 {
			t.Log("Expected validation errors but got none")
		}
		assert.NotEmpty(t, errors, "Expected validation errors")
		assert.Contains(t, errors, "User ID is required", "Expected error about missing user ID")
	})
	
	t.Run("InvalidOrder_NoDeliveryAddress", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 9.99)
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create order with missing DeliveryAddress
		order := models.Order{
			UserID:       user.ID,
			DeliveryDate: time.Now().Add(24 * time.Hour),
			OrderItems:   []models.OrderItem{orderItem},
		}
		
		// Validate order
		errors := order.ValidateOrder()
		assert.NotEmpty(t, errors, "Expected validation errors")
		assert.Contains(t, errors, "Delivery address is required", "Expected error about missing delivery address")
	})
	
	t.Run("InvalidOrder_NoOrderItems", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create order with no items
		order := models.Order{
			UserID:          user.ID,
			DeliveryAddress: "123 Test St, Test City, Test State 12345",
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			OrderItems:      []models.OrderItem{},
		}
		
		// Validate order
		errors := order.ValidateOrder()
		assert.NotEmpty(t, errors, "Expected validation errors")
		assert.Contains(t, errors, "Order must contain at least one item", "Expected error about missing order items")
	})
}

func TestOrderStatusTransitions(t *testing.T) {
	db := tests.SetupTestSuite(t)
	
	t.Run("TransitionFromPendingToPaid", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 9.99)
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create an order with Pending status
		order := models.Order{
			UserID:          user.ID,
			Status:          models.OrderStatusPending,
			DeliveryAddress: "123 Test St, Test City, Test State 12345",
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			OrderItems:      []models.OrderItem{orderItem},
		}
		
		// Save the order
		db.Create(&order)
		
		// Update status from Pending to Paid
		order.Status = models.OrderStatusPaid
		err := db.Save(&order).Error
		
		assert.Nil(t, err, "Expected no error when updating status from Pending to Paid")
		assert.Equal(t, models.OrderStatusPaid, order.Status, "Expected order status to be Paid")
	})
	
	t.Run("TransitionFromPendingToDelivered_Skip_Status", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create test user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		// Create a meal
		meal := tests.CreateTestMeal(db, "Test Meal", 9.99)
		
		// Create order item
		orderItem := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal.ID,
			Quantity: 2,
			Price:    meal.Price,
			Name:     meal.Name,
		}
		
		// Create an order with Pending status
		order := models.Order{
			UserID:          user.ID,
			Status:          models.OrderStatusPending,
			DeliveryAddress: "123 Test St, Test City, Test State 12345", 
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			OrderItems:      []models.OrderItem{orderItem},
		}
		
		// Save the order
		db.Create(&order)
		
		// Attempt to skip statuses (Pending -> Delivered)
		// This would normally be caught by the handler validation, but the DB would allow it
		order.Status = models.OrderStatusDelivered
		err := db.Save(&order).Error
		
		assert.Nil(t, err, "Database should allow status change")
		assert.Equal(t, models.OrderStatusDelivered, order.Status, "Status should be updated at database level")
		
		// Note: In real application, this would be prevented by handlers using isValidStatusTransition
	})
}

func TestOrderCalculations(t *testing.T) {
	db := tests.SetupTestSuite(t)
	
	t.Run("CalculateTotalAmount", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create an order with multiple items
		user := tests.CreateTestUser(db, models.UserTypeCustomer)
		
		meal1 := tests.CreateTestMeal(db, "Meal 1", 9.99)
		meal2 := tests.CreateTestMeal(db, "Meal 2", 14.99)
		
		orderItem1 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal1.ID,
			Quantity: 2,
			Price:    meal1.Price,
			Name:     meal1.Name,
		}
		
		orderItem2 := models.OrderItem{
			ItemType: models.OrderItemTypeMeal,
			ItemID:   meal2.ID,
			Quantity: 1,
			Price:    meal2.Price,
			Name:     meal2.Name,
		}
		
		order := models.Order{
			UserID:          user.ID,
			Status:          models.OrderStatusPending,
			DeliveryAddress: "123 Test St, Test City, Test State 12345",
			DeliveryDate:    time.Now().Add(24 * time.Hour),
			OrderItems:      []models.OrderItem{orderItem1, orderItem2},
		}
		
		// Calculate total
		order.CalculateTotalAmount()
		
		// Expected: (9.99 * 2) + (14.99 * 1) = 19.98 + 14.99 = 34.97
		expectedTotal := 9.99*2 + 14.99
		assert.Equal(t, expectedTotal, order.TotalAmount, "Expected total amount to be 34.97")
	})
}
