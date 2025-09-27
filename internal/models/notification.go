package models

import "time"

type Subscription struct {
	ID        string    `json:"id" db:"id"`
	AccountID string    `json:"account_id" db:"account_id"`
	Kind      string    `json:"kind" db:"kind"` // push | sms
	Endpoint  string    `json:"endpoint" db:"endpoint"`
	Metadata  string    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
