package models

import (
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

type PaymentMethod string

const (
	PaymentMethodMobileMoney PaymentMethod = "mobile_money"
	PaymentMethodCash        PaymentMethod = "cash"
	PaymentMethodCard        PaymentMethod = "card"
)

type Payment struct {
	ID            string        `json:"id" db:"id"`
	OrderID       string        `json:"order_id" db:"order_id"`
	Amount        float64       `json:"amount" db:"amount"`
	Method        PaymentMethod `json:"method" db:"method"`
	Status        PaymentStatus `json:"status" db:"status"`
	TransactionID string        `json:"transaction_id,omitempty" db:"transaction_id"`
	PhoneNumber   string        `json:"phone_number,omitempty" db:"phone_number"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

type ProcessPaymentRequest struct {
	OrderID     string        `json:"order_id" binding:"required"`
	Method      PaymentMethod `json:"method" binding:"required"`
	PhoneNumber string        `json:"phone_number,omitempty"`
}

type PaymentResponse struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"order_id"`
	Amount        float64       `json:"amount"`
	Method        PaymentMethod `json:"method"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id,omitempty"`
	Message       string        `json:"message,omitempty"`
	CheckoutURL   string        `json:"checkout_url,omitempty"`
}
