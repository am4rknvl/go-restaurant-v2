package models

import (
	"time"
)

type Account struct {
	ID        string    `json:"id" db:"id"`
	PhoneNumber string  `json:"phone_number" db:"phone_number"`
	Balance   float64   `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateAccountRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type AccountBalanceResponse struct {
	AccountID string  `json:"account_id"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
}
