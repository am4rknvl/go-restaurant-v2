package database

import (
	"log"

	"restaurant-system/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewGorm creates a GORM DB for features that use ORM (menu mgmt)
func NewGorm(pgURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&models.MenuCategory{}, &models.MenuItemGorm{}, &models.MenuVariant{}, &models.MenuAddon{},
		&models.UserRole{}, &models.InventoryItem{}, &models.InventoryAdjustment{},
		&models.StaffAssignment{}, &models.OrderAudit{}, &models.Discount{}, &models.DiscountUsage{},
		&models.LoyaltyAccount{}, &models.LoyaltyTransaction{}, &models.Restaurant{},
		&models.TableState{}, &models.WaitlistEntry{}, &models.PaymentTip{},
	); err != nil {
		return nil, err
	}
	log.Println("GORM AutoMigrate completed for menu models")
	return db, nil
}
