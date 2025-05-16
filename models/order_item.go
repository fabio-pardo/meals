package models

import (
	"time"
)

// OrderItemType represents the type of item ordered (meal or menu)
type OrderItemType string

const (
	OrderItemTypeMeal OrderItemType = "meal"
	OrderItemTypeMenu OrderItemType = "menu"
)

// OrderItem represents a single item in an order
type OrderItem struct {
	ID        uint          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"autoUpdateTime;not null"`
	OrderID   uint          `json:"order_id" gorm:"not null"`
	Order     Order         `json:"-" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"`
	ItemType  OrderItemType `json:"item_type" gorm:"type:varchar(10);not null"`
	ItemID    uint          `json:"item_id" gorm:"not null"` // ID of either Meal or Menu
	Quantity  int           `json:"quantity" gorm:"not null;default:1"`
	Price     float64       `json:"price" gorm:"not null"` // Price at the time of order (copy of the current price)
	Name      string        `json:"name" gorm:"not null"`  // Name at the time of order (copy of the current name)
	Notes     string        `json:"notes" gorm:"type:text"`
	MealID    *uint         `json:"meal_id,omitempty" gorm:"-"` // Used for validation and association, not stored in DB
	MenuID    *uint         `json:"menu_id,omitempty" gorm:"-"` // Used for validation and association, not stored in DB
}

// ValidateOrderItem validates the order item data
func (oi *OrderItem) ValidateOrderItem() []string {
	var errors []string

	if oi.OrderID == 0 {
		errors = append(errors, "Order ID is required")
	}

	if oi.ItemType != OrderItemTypeMeal && oi.ItemType != OrderItemTypeMenu {
		errors = append(errors, "Item type must be either 'meal' or 'menu'")
	}

	if oi.ItemID == 0 {
		errors = append(errors, "Item ID is required")
	}

	if oi.Quantity <= 0 {
		errors = append(errors, "Quantity must be greater than 0")
	}

	if oi.Price <= 0 {
		errors = append(errors, "Price must be greater than 0")
	}

	if oi.Name == "" {
		errors = append(errors, "Name is required")
	}

	return errors
}

// SetFromMeal sets the OrderItem fields based on a Meal
func (oi *OrderItem) SetFromMeal(meal *Meal) {
	oi.ItemType = OrderItemTypeMeal
	oi.ItemID = meal.ID
	oi.Price = meal.Price
	oi.Name = meal.Name
}

// SetFromMenu sets the OrderItem fields based on a Menu
func (oi *OrderItem) SetFromMenu(menu *Menu) {
	oi.ItemType = OrderItemTypeMenu
	oi.ItemID = menu.ID
	oi.Name = menu.Name
	// Price would need to be calculated based on the included meals
}
