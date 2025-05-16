package handlers

import (
	"meals/models"
	"meals/store"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	DeliveryAddress string                `json:"delivery_address" binding:"required"`
	DeliveryDate    time.Time             `json:"delivery_date" binding:"required"`
	DeliveryNotes   string                `json:"delivery_notes"`
	PaymentMethod   string                `json:"payment_method"`
	Items           []CreateOrderItemData `json:"items" binding:"required,min=1"`
}

// CreateOrderItemData represents the data for each item in an order
type CreateOrderItemData struct {
	ItemType models.OrderItemType `json:"item_type" binding:"required,oneof=meal menu"`
	ItemID   uint                 `json:"item_id" binding:"required"`
	Quantity int                  `json:"quantity" binding:"required,min=1"`
	Notes    string               `json:"notes"`
}

// OrderResponse represents the response for order operations
type OrderResponse struct {
	Order models.Order `json:"order"`
}

// CreateOrderHandler handles the creation of a new order
func CreateOrderHandler(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, ValidationError("Invalid request data", err.Error()))
		return
	}

	// Get current user from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "User must be logged in to create an order",
		})
		return
	}

	order := models.Order{
		UserID:          userID.(uint),
		DeliveryAddress: req.DeliveryAddress,
		DeliveryDate:    req.DeliveryDate,
		DeliveryNotes:   req.DeliveryNotes,
		Status:          models.OrderStatusPending,
		PaymentMethod:   req.PaymentMethod,
	}

	// Run in transaction
	err := store.WithTransaction(c, func(tx *gorm.DB) error {
		// Validate and create order items
		for _, itemData := range req.Items {
			orderItem, err := createOrderItem(tx, itemData, &order)
			if err != nil {
				return err
			}
			order.OrderItems = append(order.OrderItems, *orderItem)
		}

		// Calculate total amount
		order.CalculateTotalAmount()

		// Validate entire order
		if validationErrors := order.ValidateOrder(); len(validationErrors) > 0 {
			return ValidationErrorType{
				Message: "Order validation failed",
				Details: validationErrors,
			}
		}

		// Create the order
		if err := tx.Create(&order).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to create order",
				Details: err.Error(),
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusCreated, OrderResponse{
		Order: order,
	})
}

// GetOrderHandler handles fetching a single order by ID
func GetOrderHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid order ID"))
		return
	}

	// Get current user from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "User must be logged in to view orders",
		})
		return
	}

	var order models.Order
	db := store.GetTxFromContext(c)

	// Find order with its items
	result := db.Preload("OrderItems").First(&order, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			HandleAppError(c, NotFoundErrorType{Resource: "Order"})
			return
		}
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch order",
			Details: result.Error.Error(),
		})
		return
	}

	// Ensure user can only view their own orders
	if order.UserID != userID.(uint) {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusForbidden,
			Code:    ErrForbidden,
			Message: "You don't have permission to view this order",
		})
		return
	}

	c.JSON(http.StatusOK, OrderResponse{
		Order: order,
	})
}

// ListOrdersHandler handles fetching all orders for the current user
func ListOrdersHandler(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "User must be logged in to list orders",
		})
		return
	}

	var orders []models.Order
	db := store.GetTxFromContext(c)

	// Find all orders for this user
	if err := db.Where("user_id = ?", userID).Preload("OrderItems").Find(&orders).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch orders",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

// UpdateOrderStatusHandler handles updating the status of an order
func UpdateOrderStatusHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid order ID"))
		return
	}

	var statusUpdate struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		RespondWithError(c, ValidationError("Invalid status update", err.Error()))
		return
	}

	// Get current user from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "User must be logged in to update orders",
		})
		return
	}

	// Update in transaction
	err = store.WithTransaction(c, func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.First(&order, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Order"}
			}
			return DatabaseErrorType{
				Message: "Failed to fetch order",
				Details: err.Error(),
			}
		}

		// Ensure user can only update their own orders
		if order.UserID != userID.(uint) {
			return ValidationErrorType{
				Message: "You don't have permission to update this order",
			}
		}

		// Only allow certain status transitions
		if !isValidStatusTransition(order.Status, statusUpdate.Status) {
			return ValidationErrorType{
				Message: "Invalid status transition",
				Details: "Cannot change status from " + string(order.Status) + " to " + string(statusUpdate.Status),
			}
		}

		// Update the status
		order.Status = statusUpdate.Status
		if err := tx.Save(&order).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to update order status",
				Details: err.Error(),
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	var updatedOrder models.Order
	if err := store.DB.Preload("OrderItems").First(&updatedOrder, id).Error; err != nil {
		HandleAppError(c, DatabaseErrorType{
			Message: "Failed to fetch updated order",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, OrderResponse{
		Order: updatedOrder,
	})
}

// CancelOrderHandler handles cancelling an order
func CancelOrderHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, BadRequestError("Invalid order ID"))
		return
	}

	// Get current user from context
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusUnauthorized,
			Code:    ErrUnauthorized,
			Message: "User must be logged in to cancel orders",
		})
		return
	}

	// Cancel in transaction
	err = store.WithTransaction(c, func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.First(&order, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NotFoundErrorType{Resource: "Order"}
			}
			return DatabaseErrorType{
				Message: "Failed to fetch order",
				Details: err.Error(),
			}
		}

		// Ensure user can only cancel their own orders
		if order.UserID != userID.(uint) {
			return ValidationErrorType{
				Message: "You don't have permission to cancel this order",
			}
		}

		// Can only cancel if it's in certain states
		if !canCancelOrder(order.Status) {
			return ValidationErrorType{
				Message: "Cannot cancel order",
				Details: "Order in " + string(order.Status) + " status cannot be cancelled",
			}
		}

		// Set status to cancelled
		order.Status = models.OrderStatusCancelled
		if err := tx.Save(&order).Error; err != nil {
			return DatabaseErrorType{
				Message: "Failed to cancel order",
				Details: err.Error(),
			}
		}

		return nil
	})

	if HandleAppError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
	})
}

// Helper functions

// createOrderItem validates and creates an order item
func createOrderItem(tx *gorm.DB, itemData CreateOrderItemData, order *models.Order) (*models.OrderItem, error) {
	orderItem := models.OrderItem{
		ItemType: itemData.ItemType,
		ItemID:   itemData.ItemID,
		Quantity: itemData.Quantity,
		Notes:    itemData.Notes,
	}

	// Fetch and validate the referenced item based on type
	if itemData.ItemType == models.OrderItemTypeMeal {
		var meal models.Meal
		if err := tx.First(&meal, itemData.ItemID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, NotFoundErrorType{Resource: "Meal"}
			}
			return nil, DatabaseErrorType{
				Message: "Failed to fetch meal",
				Details: err.Error(),
			}
		}

		orderItem.SetFromMeal(&meal)
	} else if itemData.ItemType == models.OrderItemTypeMenu {
		var menu models.Menu
		if err := tx.Preload("MenuMeals").Preload("MenuMeals.Meal").First(&menu, itemData.ItemID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, NotFoundErrorType{Resource: "Menu"}
			}
			return nil, DatabaseErrorType{
				Message: "Failed to fetch menu",
				Details: err.Error(),
			}
		}

		orderItem.SetFromMenu(&menu)

		// Calculate menu price by summing the prices of all included meals
		var totalPrice float64
		for _, menuMeal := range menu.MenuMeals {
			totalPrice += menuMeal.Meal.Price
		}
		orderItem.Price = totalPrice
	}

	// Validate the order item
	if validationErrors := orderItem.ValidateOrderItem(); len(validationErrors) > 0 {
		return nil, ValidationErrorType{
			Message: "Order item validation failed",
			Details: validationErrors,
		}
	}

	return &orderItem, nil
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(current, target models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending: {
			models.OrderStatusPaid,
			models.OrderStatusCancelled,
		},
		models.OrderStatusPaid: {
			models.OrderStatusPreparing,
			models.OrderStatusCancelled,
		},
		models.OrderStatusPreparing: {
			models.OrderStatusDelivering,
		},
		models.OrderStatusDelivering: {
			models.OrderStatusDelivered,
		},
		// No transitions from delivered or cancelled states
	}

	if transitions, exists := validTransitions[current]; exists {
		for _, validTarget := range transitions {
			if target == validTarget {
				return true
			}
		}
	}

	return false
}

// canCancelOrder checks if an order can be cancelled based on its status
func canCancelOrder(status models.OrderStatus) bool {
	// Orders can only be cancelled if they are pending or paid
	return status == models.OrderStatusPending || status == models.OrderStatusPaid
}
