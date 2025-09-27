package models

import "time"

type Reservation struct {
	ID          string    `json:"id" db:"id"`
	AccountID   string    `json:"account_id" db:"account_id"`
	TableID     string    `json:"table_id" db:"table_id"`
	PartySize   int       `json:"party_size" db:"party_size"`
	ReservedFor time.Time `json:"reserved_for" db:"reserved_for"`
	Status      string    `json:"status" db:"status"`
	Notes       string    `json:"notes,omitempty" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
