package models

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          string      `json:"id" db:"id"`
	CustomerID  string      `json:"customer_id" db:"customer_id"`
	SessionID   string      `json:"session_id,omitempty" db:"session_id"`
	Items       []OrderItem `json:"items" db:"items"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	Status      OrderStatus `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID                  string  `json:"id" db:"id"`
	OrderID             string  `json:"order_id" db:"order_id"`
	MenuItemID          string  `json:"menu_item_id" db:"menu_item_id"`
	Name                string  `json:"name" db:"name"`
	Price               float64 `json:"price" db:"price"`
	Quantity            int     `json:"quantity" db:"quantity"`
	TotalPrice          float64 `json:"total_price" db:"total_price"`
	SpecialInstructions string  `json:"special_instructions,omitempty" db:"special_instructions"`
}

type MenuItem struct {
	ID            string  `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Description   string  `json:"description" db:"description"`
	Price         float64 `json:"price" db:"price"`
	Category      string  `json:"category" db:"category"`
	Available     bool    `json:"available" db:"available"`
	ImageURL      string  `json:"image_url,omitempty" db:"image_url"`
	SpecialNotes  string  `json:"special_notes,omitempty" db:"special_notes"`
	NameAm        string  `json:"name_am,omitempty" db:"name_am"`
	DescriptionAm string  `json:"description_am,omitempty" db:"description_am"`
}

type Favorite struct {
	ID         string `json:"id" db:"id"`
	AccountID  string `json:"account_id" db:"account_id"`
	MenuItemID string `json:"menu_item_id" db:"menu_item_id"`
}

type Review struct {
	ID         string `json:"id" db:"id"`
	AccountID  string `json:"account_id" db:"account_id"`
	MenuItemID string `json:"menu_item_id" db:"menu_item_id"`
	Rating     int    `json:"rating" db:"rating"`
	Comment    string `json:"comment,omitempty" db:"comment"`
}

type CreateOrderRequest struct {
	CustomerID string            `json:"customer_id" binding:"required"`
	SessionID  string            `json:"session_id,omitempty"`
	Items      []CreateOrderItem `json:"items" binding:"required"`
}

type CreateOrderItem struct {
	MenuItemID          string `json:"menu_item_id" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,min=1"`
	SpecialInstructions string `json:"special_instructions,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}
