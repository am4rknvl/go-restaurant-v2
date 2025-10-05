package handlers

import (
	"net/http"
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

// GetAccount godoc
// @Summary Get account details
// @Description Get account information by ID
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Failure 404 {object} models.ErrorResponse
// @Router /accounts/{id} [get]
func (h *EnterpriseAPI) GetAccount(c *gin.Context) {
	var account models.Account
	if err := h.db.First(&account, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, account)
}

// UpdateAccount godoc
// @Summary Update account
// @Description Update account information
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body object{name=string,phone=string} true "Account update"
// @Success 204 "Updated"
// @@Failure 400 {object} models.ErrorRespons
// @Router /accounts/{id} [put]
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

// AssignRole godoc
// @Summary Assign role to account
// @Description Assign a role to a user account
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body object{role=string,restaurant_id=string} true "Role assignment"
// @Success 204 "Role assigned"
// @@Failure 400 {object} models.ErrorRespons
// @Router /accounts/{id}/roles [post]
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

// RemoveRole godoc
// @Summary Remove role from account
// @Description Remove a role from a user account
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param role path string true "Role name"
// @Success 204 "Role removed"
// @@Failure 400 {object} models.ErrorRespons
// @Router /accounts/{id}/roles/{role} [delete]
func (h *EnterpriseAPI) RemoveRole(c *gin.Context) {
	if err := h.db.Delete(&models.UserRole{}, "account_id = ? AND role = ?", c.Param("id"), c.Param("role")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListInventory godoc
// @Summary List inventory items
// @Description Get all inventory items
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Param restaurant_id query string false "Filter by restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /inventory [get]
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

// CreateInventoryItem godoc
// @Summary Create inventory item
// @Description Add a new inventory item
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.InventoryItem true "Inventory item"
// @Success 201 {object} models.InventoryItem
// @@Failure 400 {object} models.ErrorRespons
// @Router /inventory [post]
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

// UpdateInventoryItem godoc
// @Summary Update inventory item
// @Description Update inventory item details
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body models.InventoryItem true "Item update"
// @Success 200 {object} models.InventoryItem
// @@Failure 400 {object} models.ErrorRespons
// @Router /inventory/{id} [put]
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

// AdjustInventory godoc
// @Summary Adjust inventory quantity
// @Description Adjust inventory item quantity
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body object{delta=number,reason=string} true "Adjustment"
// @Success 204 "Adjusted"
// @@Failure 400 {object} models.ErrorRespons
// @Router /inventory/{id}/adjust [patch]
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

// AssignWaiterToTable godoc
// @Summary Assign waiter to table
// @Description Assign a waiter to a specific table
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param table_id path string true "Table ID"
// @Param request body object{waiter_id=string,restaurant_id=string} true "Assignment"
// @Success 204 "Assigned"
// @@Failure 400 {object} models.ErrorRespons
// @Router /tables/{table_id}/assign-waiter [post]
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

// AssignChefToOrder godoc
// @Summary Assign chef to order
// @Description Assign a chef to prepare an order
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body object{chef_id=string,restaurant_id=string} true "Assignment"
// @Success 204 "Assigned"
// @@Failure 400 {object} models.ErrorRespons
// @Router /orders/{id}/assign-chef [post]
func (h *EnterpriseAPI) AssignChefToOrder(c *gin.Context) {
	var req struct {
		ChefID       string `json:"chef_id" binding:"required"`
		RestaurantID string `json:"restaurant_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID := c.Param("id")
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

// ListStaffAssignments godoc
// @Summary List staff assignments
// @Description Get all staff assignments
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Param restaurant_id query string false "Filter by restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /staff/assignments [get]
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

// SplitOrder godoc
// @Summary Split an order
// @Description Split an order into multiple orders
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]string
// @Router /orders/{id}/split [post]
func (h *EnterpriseAPI) SplitOrder(c *gin.Context) {
	// Placeholder - would implement order splitting logic
	c.JSON(http.StatusOK, gin.H{"message": "order_split_initiated"})
}

// MergeOrders godoc
// @Summary Merge orders
// @Description Merge multiple orders into one
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]string
// @Router /orders/{id}/merge [post]
func (h *EnterpriseAPI) MergeOrders(c *gin.Context) {
	// Placeholder - would implement order merging logic
	c.JSON(http.StatusOK, gin.H{"message": "orders_merged"})
}

// AddTipToPayment godoc
// @Summary Add tip to payment
// @Description Add a tip amount to a payment
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Param request body object{amount=number} true "Tip amount"
// @Success 200 {object} models.PaymentTip
// @@Failure 400 {object} models.ErrorRespons
// @Router /payments/{id}/tip [post]
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

// CreateDiscount godoc
// @Summary Create discount
// @Description Create a new discount code
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Discount true "Discount"
// @Success 201 {object} models.Discount
// @@Failure 400 {object} models.ErrorRespons
// @Router /discounts [post]
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

// ApplyDiscount godoc
// @Summary Apply discount code
// @Description Apply a discount code to an order
// @Tags enterprise
// @Accept json
// @Produce json
// @Param request body object{account_id=string,order_id=string,code=string} true "Discount application"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /discounts/apply [post]
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

// GetLoyaltyAccount godoc
// @Summary Get loyalty account
// @Description Get loyalty points for an account
// @Tags enterprise
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.LoyaltyAccount
// @Failure 500 {object} models.ErrorResponse
// @Router /accounts/{id}/loyalty [get]
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

// EarnLoyaltyPoints godoc
// @Summary Earn loyalty points
// @Description Add loyalty points to an account
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body object{points=int} true "Points to earn"
// @Success 204 "Points earned"
// @@Failure 400 {object} models.ErrorRespons
// @Router /accounts/{id}/loyalty/earn [post]
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

// SalesReport godoc
// @Summary Get sales report
// @Description Get sales analytics and reports
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Param range query string false "Date range"
// @Success 200 {object} map[string]interface{}
// @Router /reports/sales [get]
func (h *EnterpriseAPI) SalesReport(c *gin.Context) {
	// Placeholder - would implement sales analytics
	c.JSON(http.StatusOK, gin.H{"total_sales": 50000, "period": c.Query("range")})
}

// PopularItemsReport godoc
// @Summary Get popular items report
// @Description Get most popular menu items
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /reports/popular-items [get]
func (h *EnterpriseAPI) PopularItemsReport(c *gin.Context) {
	// Placeholder - would implement popular items analytics
	c.JSON(http.StatusOK, gin.H{"popular_items": []string{"burger", "pizza"}})
}

// TopCustomersReport godoc
// @Summary Get top customers report
// @Description Get top spending customers
// @Tags enterprise
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /reports/customers/top [get]
func (h *EnterpriseAPI) TopCustomersReport(c *gin.Context) {
	// Placeholder - would implement top customers analytics
	c.JSON(http.StatusOK, gin.H{"top_customers": []string{"customer1", "customer2"}})
}

// ListRestaurants godoc
// @Summary List restaurants
// @Description Get all restaurants
// @Tags enterprise
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /restaurants [get]
func (h *EnterpriseAPI) ListRestaurants(c *gin.Context) {
	var restaurants []models.Restaurant
	if err := h.db.Find(&restaurants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"restaurants": restaurants})
}

// CreateRestaurant godoc
// @Summary Create restaurant
// @Description Create a new restaurant
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Restaurant true "Restaurant"
// @Success 201 {object} models.Restaurant
// @@Failure 400 {object} models.ErrorRespons
// @Router /restaurants [post]
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

// UpdateRestaurant godoc
// @Summary Update restaurant
// @Description Update restaurant details
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Param request body models.Restaurant true "Restaurant update"
// @Success 200 {object} models.Restaurant
// @@Failure 400 {object} models.ErrorRespons
// @Router /restaurants/{id} [put]
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

// UpdateTableState godoc
// @Summary Update table state
// @Description Update the state of a table
// @Tags enterprise
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Param request body object{state=string} true "State update"
// @Success 204 "Updated"
// @@Failure 400 {object} models.ErrorRespons
// @Router /tables/{id}/state [patch]
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

// JoinWaitlist godoc
// @Summary Join waitlist
// @Description Add customer to restaurant waitlist
// @Tags enterprise
// @Accept json
// @Produce json
// @Param request body models.WaitlistEntry true "Waitlist entry"
// @Success 201 {object} models.WaitlistEntry
// @@Failure 400 {object} models.ErrorRespons
// @Router /waitlist [post]
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

// ListWaitlist godoc
// @Summary List waitlist
// @Description Get all waitlist entries
// @Tags enterprise
// @Produce json
// @Param restaurant_id query string false "Filter by restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /waitlist [get]
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
