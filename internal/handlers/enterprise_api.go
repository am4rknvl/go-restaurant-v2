package handlers

import (
	"net/http"
	"strconv"
	"time"

	"restaurant-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnterpriseAPI struct {
	db *gorm.DB
	ws interface{ Broadcast(v interface{}) }
}

func NewEnterpriseAPI(db *gorm.DB, ws interface{ Broadcast(v interface{}) }) *EnterpriseAPI {
	return &EnterpriseAPI{db: db, ws: ws}
}

// User Profiles & Role Management
func (h *EnterpriseAPI) GetAccount(c *gin.Context) {
	var account models.Account
	if err := h.db.First(&account, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *EnterpriseAPI) UpdateAccount(c *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Model(&models.Account{}).Where("id = ?", c.Param("id")).Updates(req).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseAPI) AssignRole(c *gin.Context) {
	var req struct {
		Role         string  `json:"role" binding:"required"`
		RestaurantID *string `json:"restaurant_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := models.UserRole{
		ID:           uuid.New().String(),
		AccountID:    c.Param("id"),
		Role:         req.Role,
		RestaurantID: req.RestaurantID,
	}
	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseAPI) RemoveRole(c *gin.Context) {
	if err := h.db.Delete(&models.UserRole{}, "account_id = ? AND role = ?", c.Param("id"), c.Param("role")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Inventory Management
func (h *EnterpriseAPI) ListInventory(c *gin.Context) {
	var items []models.InventoryItem
	q := h.db.Model(&models.InventoryItem{})
	if rid := c.Query("restaurant_id"); rid != "" {
		q = q.Where("restaurant_id = ?", rid)
	}
	if err := q.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *EnterpriseAPI) CreateInventoryItem(c *gin.Context) {
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	if err := h.db.Create(&item).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *EnterpriseAPI) UpdateInventoryItem(c *gin.Context) {
	var item models.InventoryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item.ID = c.Param("id")
	if err := h.db.Model(&models.InventoryItem{ID: item.ID}).Updates(item).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *EnterpriseAPI) AdjustInventory(c *gin.Context) {
	var req struct {
		Delta  float64 `json:"delta" binding:"required"`
		Reason string  `json:"reason"`
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

	if err := tx.Model(&models.InventoryItem{}).Where("id = ?", c.Param("id")).Update("qty", gorm.Expr("qty + ?", req.Delta)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adj := models.InventoryAdjustment{
		ID:     uuid.New().String(),
		ItemID: c.Param("id"),
		Delta:  req.Delta,
		Reason: req.Reason,
		UserID: c.GetString("account_id"),
	}
	if err := tx.Create(&adj).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

// Staff Assignment
func (h *EnterpriseAPI) AssignWaiterToTable(c *gin.Context) {
	var req struct {
		WaiterID     string `json:"waiter_id" binding:"required"`
		RestaurantID string `json:"restaurant_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	tableID := c.Param("table_id")
	assignment := models.StaffAssignment{
		ID:           uuid.New().String(),
		RestaurantID: req.RestaurantID,
		StaffID:      req.WaiterID,
		TableID:      &tableID,
		AssignType:   "waiter",
	}
	if err := h.db.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseAPI) AssignChefToOrder(c *gin.Context) {
	var req struct {
		ChefID       string `json:"chef_id" binding:"required"`
		RestaurantID string `json:"restaurant_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	orderID := c.Param("order_id")
	assignment := models.StaffAssignment{
		ID:           uuid.New().String(),
		RestaurantID: req.RestaurantID,
		StaffID:      req.ChefID,
		OrderID:      &orderID,
		AssignType:   "chef",
	}
	if err := h.db.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseAPI) ListStaffAssignments(c *gin.Context) {
	var assignments []models.StaffAssignment
	q := h.db.Model(&models.StaffAssignment{})
	if rid := c.Query("restaurant_id"); rid != "" {
		q = q.Where("restaurant_id = ?", rid)
	}
	if err := q.Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"assignments": assignments})
}

// Order Extensions
func (h *EnterpriseAPI) SplitOrder(c *gin.Context) {
	// Placeholder - would implement order splitting logic
	c.JSON(http.StatusOK, gin.H{"message": "order_split_initiated"})
}

func (h *EnterpriseAPI) MergeOrders(c *gin.Context) {
	// Placeholder - would implement order merging logic
	c.JSON(http.StatusOK, gin.H{"message": "orders_merged"})
}

func (h *EnterpriseAPI) AddTipToPayment(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	tip := models.PaymentTip{
		ID:        uuid.New().String(),
		PaymentID: c.Param("id"),
		Amount:    req.Amount,
	}
	if err := h.db.Create(&tip).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tip)
}

// Discounts & Loyalty
func (h *EnterpriseAPI) CreateDiscount(c *gin.Context) {
	var discount models.Discount
	if err := c.ShouldBindJSON(&discount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if discount.ID == "" {
		discount.ID = uuid.New().String()
	}
	if err := h.db.Create(&discount).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, discount)
}

func (h *EnterpriseAPI) ApplyDiscount(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id"`
		OrderID   string `json:"order_id"`
		Code      string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var discount models.Discount
	if err := h.db.Where("code = ? AND valid_from <= ? AND valid_to >= ?", req.Code, time.Now(), time.Now()).First(&discount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_or_expired_discount"})
		return
	}
	
	// Apply discount logic here
	c.JSON(http.StatusOK, gin.H{"discount_applied": true, "amount": discount.Value})
}

func (h *EnterpriseAPI) GetLoyaltyAccount(c *gin.Context) {
	var loyalty models.LoyaltyAccount
	if err := h.db.Where("account_id = ?", c.Param("id")).First(&loyalty).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			loyalty = models.LoyaltyAccount{
				ID:        uuid.New().String(),
				AccountID: c.Param("id"),
				Points:    0,
				Tier:      "bronze",
			}
			h.db.Create(&loyalty)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, loyalty)
}

func (h *EnterpriseAPI) EarnLoyaltyPoints(c *gin.Context) {
	var req struct {
		Points int `json:"points" binding:"required"`
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

	if err := tx.Model(&models.LoyaltyAccount{}).Where("account_id = ?", c.Param("id")).Update("points", gorm.Expr("points + ?", req.Points)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction := models.LoyaltyTransaction{
		ID:        uuid.New().String(),
		AccountID: c.Param("id"),
		Points:    req.Points,
		Type:      "earn",
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.Status(http.StatusNoContent)
}

// Analytics & Reporting
func (h *EnterpriseAPI) SalesReport(c *gin.Context) {
	// Placeholder - would implement sales analytics
	c.JSON(http.StatusOK, gin.H{"total_sales": 50000, "period": c.Query("range")})
}

func (h *EnterpriseAPI) PopularItemsReport(c *gin.Context) {
	// Placeholder - would implement popular items analytics
	c.JSON(http.StatusOK, gin.H{"popular_items": []string{"burger", "pizza"}})
}

func (h *EnterpriseAPI) TopCustomersReport(c *gin.Context) {
	// Placeholder - would implement top customers analytics
	c.JSON(http.StatusOK, gin.H{"top_customers": []string{"customer1", "customer2"}})
}

// Restaurant Management
func (h *EnterpriseAPI) ListRestaurants(c *gin.Context) {
	var restaurants []models.Restaurant
	if err := h.db.Find(&restaurants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"restaurants": restaurants})
}

func (h *EnterpriseAPI) CreateRestaurant(c *gin.Context) {
	var restaurant models.Restaurant
	if err := c.ShouldBindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if restaurant.ID == "" {
		restaurant.ID = uuid.New().String()
	}
	if err := h.db.Create(&restaurant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, restaurant)
}

func (h *EnterpriseAPI) UpdateRestaurant(c *gin.Context) {
	var restaurant models.Restaurant
	if err := c.ShouldBindJSON(&restaurant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	restaurant.ID = c.Param("id")
	if err := h.db.Model(&models.Restaurant{ID: restaurant.ID}).Updates(restaurant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, restaurant)
}

// Table State & Waitlist
func (h *EnterpriseAPI) UpdateTableState(c *gin.Context) {
	var req struct {
		State string `json:"state" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	tableState := models.TableState{
		ID:      uuid.New().String(),
		TableID: c.Param("id"),
		State:   req.State,
	}
	if err := h.db.Create(&tableState).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	h.ws.Broadcast(gin.H{"type": "table.state_changed", "table_id": c.Param("id"), "state": req.State})
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseAPI) JoinWaitlist(c *gin.Context) {
	var entry models.WaitlistEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	
	// Get current position
	var count int64
	h.db.Model(&models.WaitlistEntry{}).Where("restaurant_id = ? AND status = 'waiting'", entry.RestaurantID).Count(&count)
	entry.Position = int(count) + 1
	
	if err := h.db.Create(&entry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	h.ws.Broadcast(gin.H{"type": "waitlist.joined", "restaurant_id": entry.RestaurantID, "position": entry.Position})
	c.JSON(http.StatusCreated, entry)
}

func (h *EnterpriseAPI) ListWaitlist(c *gin.Context) {
	var entries []models.WaitlistEntry
	q := h.db.Model(&models.WaitlistEntry{}).Where("status = 'waiting'")
	if rid := c.Query("restaurant_id"); rid != "" {
		q = q.Where("restaurant_id = ?", rid)
	}
	if err := q.Order("position asc").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"waitlist": entries})
}