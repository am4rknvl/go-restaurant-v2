package models

import (
	"time"

	"gorm.io/gorm"
)

type TelebirrToken struct {
	ID          string    `json:"id" gorm:"primaryKey;type:text"`
	AccessToken string    `json:"access_token" gorm:"type:text;not null"`
	TokenType   string    `json:"token_type" gorm:"type:text;default:'Bearer'"`
	ExpiresIn   int       `json:"expires_in" gorm:"not null"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
}

type TelebirrOrder struct {
	ID             string         `json:"id" gorm:"primaryKey;type:text"`
	OrderID        string         `json:"order_id" gorm:"index;type:text;not null"`
	PrepayID       string         `json:"prepay_id" gorm:"uniqueIndex;type:text;not null"`
	MerchOrderID   string         `json:"merch_order_id" gorm:"type:text;not null"`
	Amount         float64        `json:"amount" gorm:"not null"`
	Currency       string         `json:"currency" gorm:"type:text;default:'ETB'"`
	Subject        string         `json:"subject" gorm:"type:text"`
	Body           string         `json:"body" gorm:"type:text"`
	NotifyURL      string         `json:"notify_url" gorm:"type:text"`
	ReturnURL      string         `json:"return_url" gorm:"type:text"`
	TimeoutExpress string         `json:"timeout_express" gorm:"type:text;default:'30m'"`
	Status         string         `json:"status" gorm:"type:text;default:'pending'"`
	PaymentURL     string         `json:"payment_url" gorm:"type:text"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TelebirrNotification struct {
	ID           string    `json:"id" gorm:"primaryKey;type:text"`
	PrepayID     string    `json:"prepay_id" gorm:"index;type:text;not null"`
	MerchOrderID string    `json:"merch_order_id" gorm:"type:text;not null"`
	TradeNo      string    `json:"trade_no" gorm:"type:text"`
	TradeStatus  string    `json:"trade_status" gorm:"type:text;not null"`
	TotalAmount  float64   `json:"total_amount" gorm:"not null"`
	Currency     string    `json:"currency" gorm:"type:text;default:'ETB'"`
	GmtPayment   time.Time `json:"gmt_payment"`
	Sign         string    `json:"sign" gorm:"type:text"`
	SignType     string    `json:"sign_type" gorm:"type:text;default:'RSA2'"`
	Processed    bool      `json:"processed" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
}

type TelebirrConfig struct {
	AppID          string `json:"app_id"`
	PrivateKey     string `json:"private_key"`
	PublicKey      string `json:"public_key"`
	NotifyURL      string `json:"notify_url"`
	ReturnURL      string `json:"return_url"`
	BaseURL        string `json:"base_url"`
	TokenURL       string `json:"token_url"`
	OrderURL       string `json:"order_url"`
	WebCheckoutURL string `json:"web_checkout_url"`
}
