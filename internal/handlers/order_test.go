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
	db.AutoMigrate(&models.Order{}, &models.OrderItem{}, &models.InventoryItem{})
	return db
}

func TestCreateOrder(t *testing.T) {
	db := setupTestDB()

	// Seed inventory
	inventory := models.InventoryItem{
		ID:           "item-1",
		RestaurantID: "rest-1",
		SKU:          "BURGER",
		Name:         "Burger",
		Qty:          10,
	}
	db.Create(&inventory)

	api := NewEnterpriseAPI(db, &mockWebSocket{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/orders", api.CreateOrder)

	orderReq := map[string]interface{}{
		"customer_id":   "cust-1",
		"restaurant_id": "rest-1",
		"items": []map[string]interface{}{
			{
				"menu_item_id": "item-1",
				"quantity":     2,
				"price":        15.99,
			},
		},
	}

	body, _ := json.Marshal(orderReq)
	req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["id"])
}

func (h *EnterpriseAPI) CreateOrder(c *gin.Context) {
	var req struct {
		CustomerID   string `json:"customer_id" binding:"required"`
		RestaurantID string `json:"restaurant_id" binding:"required"`
		Items        []struct {
			MenuItemID string  `json:"menu_item_id" binding:"required"`
			Quantity   int     `json:"quantity" binding:"required"`
			Price      float64 `json:"price" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create order
	order := models.Order{
		ID:         "order-" + req.CustomerID,
		CustomerID: req.CustomerID,
		Status:     "pending",
	}

	var total float64
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)

		// Decrement inventory
		if err := tx.Model(&models.InventoryItem{}).
			Where("id = ?", item.MenuItemID).
			Update("qty", gorm.Expr("qty - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "inventory update failed"})
			return
		}
	}

	order.TotalAmount = total
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, order)
}

type mockWebSocket struct{}

func (m *mockWebSocket) Broadcast(v interface{}) {}
