package models

import "time"

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusCompleted RefundStatus = "completed"
	RefundStatusRejected  RefundStatus = "rejected"
)

type Refund struct {
	ID        string       `json:"id" db:"id"`
	PaymentID string       `json:"payment_id" db:"payment_id"`
	Amount    float64      `json:"amount" db:"amount"`
	Reason    string       `json:"reason,omitempty" db:"reason"`
	Status    RefundStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}
