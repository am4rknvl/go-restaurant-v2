package models

import (
	"time"

	"gorm.io/gorm"
)

// H5 C2B specific models for customer-to-business payments
type TelebirrC2BOrder struct {
	ID             string         `json:"id" gorm:"primaryKey;type:text"`
	OrderID        string         `json:"order_id" gorm:"index;type:text;not null"`
	OutTradeNo     string         `json:"out_trade_no" gorm:"uniqueIndex;type:text;not null"` // Merchant order ID
	Subject        string         `json:"subject" gorm:"type:text;not null"`
	Body           string         `json:"body" gorm:"type:text"`
	TotalAmount    float64        `json:"total_amount" gorm:"not null"`
	Currency       string         `json:"currency" gorm:"type:text;default:'ETB'"`
	NotifyURL      string         `json:"notify_url" gorm:"type:text"`
	ReturnURL      string         `json:"return_url" gorm:"type:text"`
	TimeoutExpress string         `json:"timeout_express" gorm:"type:text;default:'30m'"`
	PassbackParams string         `json:"passback_params" gorm:"type:text"` // Additional params
	Status         string         `json:"status" gorm:"type:text;default:'pending'"`
	H5PayURL       string         `json:"h5_pay_url" gorm:"type:text"` // Generated H5 payment URL
	TradeNo        string         `json:"trade_no" gorm:"type:text"`   // Telebirr transaction ID
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TelebirrC2BNotification struct {
	ID             string    `json:"id" gorm:"primaryKey;type:text"`
	OutTradeNo     string    `json:"out_trade_no" gorm:"index;type:text;not null"`
	TradeNo        string    `json:"trade_no" gorm:"type:text;not null"`
	TradeStatus    string    `json:"trade_status" gorm:"type:text;not null"`
	TotalAmount    float64   `json:"total_amount" gorm:"not null"`
	Currency       string    `json:"currency" gorm:"type:text;default:'ETB'"`
	GmtPayment     time.Time `json:"gmt_payment"`
	PassbackParams string    `json:"passback_params" gorm:"type:text"`
	Sign           string    `json:"sign" gorm:"type:text"`
	SignType       string    `json:"sign_type" gorm:"type:text;default:'RSA2'"`
	Processed      bool      `json:"processed" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
}

type TelebirrC2BConfig struct {
	AppID           string `json:"app_id"`
	PrivateKey      string `json:"private_key"`
	PublicKey       string `json:"public_key"`
	NotifyURL       string `json:"notify_url"`
	ReturnURL       string `json:"return_url"`
	H5PayURL        string `json:"h5_pay_url"`        // H5 payment page URL
	UnifiedOrderURL string `json:"unified_order_url"` // Create order API
	RefundURL       string `json:"refund_url"`        // Refund API (optional; falls back to unified URL)
	QueryURL        string `json:"query_url"`         // Query API (optional; falls back to unified URL)
}

// Request structure for H5 C2B order creation
type TelebirrC2BOrderRequest struct {
	AppID      string `json:"appid" binding:"required"`
	Method     string `json:"method" binding:"required"`    // telebirr.payment.h5pay
	Format     string `json:"format" binding:"required"`    // JSON
	Charset    string `json:"charset" binding:"required"`   // utf-8
	SignType   string `json:"sign_type" binding:"required"` // RSA2
	Sign       string `json:"sign"`
	Timestamp  string `json:"timestamp" binding:"required"`
	Version    string `json:"version" binding:"required"` // 1.0
	NotifyURL  string `json:"notify_url" binding:"required"`
	BizContent string `json:"biz_content" binding:"required"` // JSON string of business content
}

// Business content for H5 C2B payment
type TelebirrC2BBizContent struct {
	OutTradeNo     string `json:"out_trade_no" binding:"required"`
	Subject        string `json:"subject" binding:"required"`
	Body           string `json:"body"`
	TotalAmount    string `json:"total_amount" binding:"required"`
	TimeoutExpress string `json:"timeout_express"`
	PassbackParams string `json:"passback_params"`
}
