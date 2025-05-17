package models

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus represents the current state of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusPreparing  OrderStatus = "preparing"
	OrderStatusDelivering OrderStatus = "delivering"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Order represents a customer order in the system
type Order struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	UserID          uint           `json:"user_id" gorm:"not null"`
	User            User           `json:"user" gorm:"foreignKey:UserID;references:ID"`
	DriverID        *uint          `json:"driver_id" gorm:"index"`
	Driver          *User          `json:"driver,omitempty" gorm:"foreignKey:DriverID;references:ID"`
	OrderItems      []OrderItem    `json:"order_items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE;"`
	Status          OrderStatus    `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	TotalAmount     float64        `json:"total_amount" gorm:"not null"`
	DeliveryAddress string         `json:"delivery_address" gorm:"type:text;not null"`
	DeliveryDate    time.Time      `json:"delivery_date" gorm:"not null"`
	DeliveryNotes   string         `json:"delivery_notes" gorm:"type:text"`
	PaymentID       string         `json:"payment_id" gorm:"type:varchar(255)"`
	PaymentMethod   string         `json:"payment_method" gorm:"type:varchar(50)"`
}

// CalculateTotalAmount calculates and updates the total amount of the order
func (o *Order) CalculateTotalAmount() {
	var total float64
	for _, item := range o.OrderItems {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = total
}

// ValidateOrder validates the order data
func (o *Order) ValidateOrder() []string {
	var errors []string

	if o.UserID == 0 {
		errors = append(errors, "User ID is required")
	}

	if o.DeliveryAddress == "" {
		errors = append(errors, "Delivery address is required")
	}

	if o.DeliveryDate.IsZero() {
		errors = append(errors, "Delivery date is required")
	}

	if len(o.OrderItems) == 0 {
		errors = append(errors, "Order must contain at least one item")
	}

	return errors
}
