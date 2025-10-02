package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	AccountID    string         `json:"account_id" gorm:"index;type:text;not null"`
	Role         string         `json:"role" gorm:"type:text;not null"`
	RestaurantID *string        `json:"restaurant_id" gorm:"index;type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type InventoryItem struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	RestaurantID string         `json:"restaurant_id" gorm:"index;type:text;not null"`
	SKU          string         `json:"sku" gorm:"type:text;not null"`
	Name         string         `json:"name" gorm:"type:text;not null"`
	Qty          float64        `json:"qty" gorm:"not null;default:0"`
	Unit         string         `json:"unit" gorm:"type:text"`
	ReorderLevel float64        `json:"reorder_level" gorm:"default:0"`
	Cost         float64        `json:"cost" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type InventoryAdjustment struct {
	ID       string    `json:"id" gorm:"primaryKey;type:text"`
	ItemID   string    `json:"item_id" gorm:"index;type:text;not null"`
	Delta    float64   `json:"delta" gorm:"not null"`
	Reason   string    `json:"reason" gorm:"type:text"`
	UserID   string    `json:"user_id" gorm:"type:text"`
	CreateAt time.Time `json:"created_at"`
}

type StaffAssignment struct {
	ID           string    `json:"id" gorm:"primaryKey;type:text"`
	RestaurantID string    `json:"restaurant_id" gorm:"index;type:text;not null"`
	StaffID      string    `json:"staff_id" gorm:"index;type:text;not null"`
	TableID      *string   `json:"table_id" gorm:"index;type:text"`
	OrderID      *string   `json:"order_id" gorm:"index;type:text"`
	AssignType   string    `json:"assign_type" gorm:"type:text;not null"`
	CreatedAt    time.Time `json:"created_at"`
}

type OrderAudit struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text"`
	OrderID   string    `json:"order_id" gorm:"index;type:text;not null"`
	Action    string    `json:"action" gorm:"type:text;not null"`
	UserID    string    `json:"user_id" gorm:"type:text"`
	Details   string    `json:"details" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

type Discount struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	Code         string         `json:"code" gorm:"uniqueIndex;type:text;not null"`
	Type         string         `json:"type" gorm:"type:text;not null"`
	Value        float64        `json:"value" gorm:"not null"`
	RestaurantID *string        `json:"restaurant_id" gorm:"index;type:text"`
	ValidFrom    time.Time      `json:"valid_from"`
	ValidTo      time.Time      `json:"valid_to"`
	UsageLimit   int            `json:"usage_limit" gorm:"default:0"`
	PerUserLimit int            `json:"per_user_limit" gorm:"default:1"`
	UsedCount    int            `json:"used_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type DiscountUsage struct {
	ID         string    `json:"id" gorm:"primaryKey;type:text"`
	DiscountID string    `json:"discount_id" gorm:"index;type:text;not null"`
	AccountID  string    `json:"account_id" gorm:"index;type:text;not null"`
	OrderID    string    `json:"order_id" gorm:"type:text"`
	Amount     float64   `json:"amount" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
}

type LoyaltyAccount struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text"`
	AccountID string    `json:"account_id" gorm:"uniqueIndex;type:text;not null"`
	Points    int       `json:"points" gorm:"default:0"`
	Tier      string    `json:"tier" gorm:"type:text;default:'bronze'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoyaltyTransaction struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text"`
	AccountID string    `json:"account_id" gorm:"index;type:text;not null"`
	Points    int       `json:"points" gorm:"not null"`
	Type      string    `json:"type" gorm:"type:text;not null"`
	OrderID   *string   `json:"order_id" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

type Restaurant struct {
	ID        string         `json:"id" gorm:"primaryKey;type:text"`
	Name      string         `json:"name" gorm:"type:text;not null"`
	Timezone  string         `json:"timezone" gorm:"type:text;default:'UTC'"`
	Currency  string         `json:"currency" gorm:"type:text;default:'USD'"`
	TaxRate   float64        `json:"tax_rate" gorm:"default:0"`
	Address   string         `json:"address" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TableState struct {
	ID           string    `json:"id" gorm:"primaryKey;type:text"`
	TableID      string    `json:"table_id" gorm:"index;type:text;not null"`
	State        string    `json:"state" gorm:"type:text;not null"`
	AssignedTo   *string   `json:"assigned_to" gorm:"type:text"`
	RestaurantID string    `json:"restaurant_id" gorm:"index;type:text;not null"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WaitlistEntry struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	RestaurantID string         `json:"restaurant_id" gorm:"index;type:text;not null"`
	Name         string         `json:"name" gorm:"type:text;not null"`
	Phone        string         `json:"phone" gorm:"type:text"`
	PartySize    int            `json:"party_size" gorm:"not null"`
	Status       string         `json:"status" gorm:"type:text;default:'waiting'"`
	Position     int            `json:"position" gorm:"default:0"`
	Notes        string         `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type PaymentTip struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text"`
	PaymentID string    `json:"payment_id" gorm:"index;type:text;not null"`
	Amount    float64   `json:"amount" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}
