package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurant-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.InventoryItem{}, &models.Restaurant{}, &models.UserRole{})
	return db
}

func TestCreateInventoryItem(t *testing.T) {
	db := setupTestDB()
	api := NewEnterpriseAPI(db, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/inventory", api.CreateInventoryItem)

	item := models.InventoryItem{
		RestaurantID: "rest1",
		SKU:          "ITEM001",
		Name:         "Test Item",
		Qty:          100,
		Unit:         "pcs",
	}
	
	jsonData, _ := json.Marshal(item)
	req, _ := http.NewRequest("POST", "/inventory", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.InventoryItem
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test Item", response.Name)
	assert.Equal(t, "ITEM001", response.SKU)
}

func TestAssignRole(t *testing.T) {
	db := setupTestDB()
	api := NewEnterpriseAPI(db, nil)
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/accounts/:id/roles", api.AssignRole)

	roleReq := struct {
		Role         string  `json:"role"`
		RestaurantID *string `json:"restaurant_id"`
	}{
		Role: "waiter",
	}
	
	jsonData, _ := json.Marshal(roleReq)
	req, _ := http.NewRequest("POST", "/accounts/user123/roles", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	
	var role models.UserRole
	db.Where("account_id = ? AND role = ?", "user123", "waiter").First(&role)
	assert.Equal(t, "waiter", role.Role)
}
