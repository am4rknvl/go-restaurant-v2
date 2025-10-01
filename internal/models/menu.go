package models

import (
	"time"

	"gorm.io/gorm"
)

// Menu Categories
type MenuCategory struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	RestaurantID string         `json:"restaurant_id" gorm:"index;type:text;not null"`
	Name         string         `json:"name" gorm:"type:text;not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Menu Items
type MenuItemGorm struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	CategoryID   string         `json:"category_id" gorm:"index;type:text;not null"`
	RestaurantID string         `json:"restaurant_id" gorm:"index;type:text;not null"`
	Name         string         `json:"name" gorm:"type:text;not null"`
	Description  string         `json:"description" gorm:"type:text"`
	Price        float64        `json:"price" gorm:"not null"`
	Available    bool           `json:"available" gorm:"default:true"`
	ImageURL     string         `json:"image_url" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Menu Variants
type MenuVariant struct {
	ID         string  `json:"id" gorm:"primaryKey;type:text"`
	ItemID     string  `json:"item_id" gorm:"index;type:text;not null"`
	Name       string  `json:"name" gorm:"type:text;not null"`
	PriceDelta float64 `json:"price_delta" gorm:"not null"`
}

// Menu Add-ons
type MenuAddon struct {
	ID         string  `json:"id" gorm:"primaryKey;type:text"`
	ItemID     string  `json:"item_id" gorm:"index;type:text;not null"`
	Name       string  `json:"name" gorm:"type:text;not null"`
	PriceDelta float64 `json:"price_delta" gorm:"not null"`
}
