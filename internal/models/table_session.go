package models

import "time"

type Table struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	QRCode    string    `json:"qr_code" db:"qr_code"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SessionStatus string

const (
	SessionStatusActive SessionStatus = "active"
	SessionStatusClosed SessionStatus = "closed"
)

type Session struct {
	ID        string        `json:"id" db:"id"`
	TableID   string        `json:"table_id" db:"table_id"`
	Customer  string        `json:"customer,omitempty" db:"customer"`
	Orders    []Order       `json:"orders,omitempty" db:"-"`
	Status    SessionStatus `json:"status" db:"status"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	ClosedAt  *time.Time    `json:"closed_at,omitempty" db:"closed_at"`
}
