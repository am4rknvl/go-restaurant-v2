package handlers

import (
	"net/http"
	"time"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"github.com/gin-gonic/gin"
)

type OrderAPI struct {
	svc *services.OrderSQLService
	hub *websocket.Hub
}

func NewOrderAPI(svc *services.OrderSQLService, hub *websocket.Hub) *OrderAPI {
	return &OrderAPI{svc: svc, hub: hub}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items for a customer
// @Tags orders
// @Accept json
// @Produce json
// @Param request body models.CreateOrderRequest true "Order request"
// @Success 201 {object} models.Order
// @@Failure 400 {object} models.ErrorRespons
// @Router /orders [post]
func (h *OrderAPI) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items := make([]services.CreateOrderItemReq, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, services.CreateOrderItemReq{MenuItemID: it.MenuItemID, Quantity: it.Quantity, SpecialInstructions: it.SpecialInstructions})
	}
	// pass optional session id when creating order
	if req.SessionID != "" {
		ord, err := h.svc.CreateOrder(c.Request.Context(), req.CustomerID, items, req.SessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if h.hub != nil {
			h.hub.Broadcast(struct {
				Type  string        `json:"type"`
				Order *models.Order `json:"order"`
			}{Type: "order_created", Order: ord})
		}
		c.JSON(http.StatusCreated, ord)
		return
	}

	ord, err := h.svc.CreateOrder(c.Request.Context(), req.CustomerID, items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// broadcast to kitchen and customer
	if h.hub != nil {
		h.hub.Broadcast(struct {
			Type  string        `json:"type"`
			Order *models.Order `json:"order"`
		}{Type: "order_created", Order: ord})
	}
	c.JSON(http.StatusCreated, ord)
}

// POST /api/v1/orders/sync
func (h *OrderAPI) SyncOrders(c *gin.Context) {
	var payload struct {
		CustomerID string                     `json:"customer_id" binding:"required"`
		SessionID  string                     `json:"session_id,omitempty"`
		Orders     [][]models.CreateOrderItem `json:"orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// convert to service types
	var batches [][]services.CreateOrderItemReq
	for _, ord := range payload.Orders {
		var items []services.CreateOrderItemReq
		for _, it := range ord {
			items = append(items, services.CreateOrderItemReq{MenuItemID: it.MenuItemID, Quantity: it.Quantity, SpecialInstructions: it.SpecialInstructions})
		}
		batches = append(batches, items)
	}
	created, err := h.svc.SyncOrders(c.Request.Context(), payload.CustomerID, batches, payload.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// broadcast newly created orders
	if h.hub != nil {
		for _, o := range created {
			h.hub.Broadcast(struct {
				Type  string        `json:"type"`
				Order *models.Order `json:"order"`
			}{Type: "order_created", Order: o})
		}
	}
	c.JSON(http.StatusCreated, gin.H{"orders": created})
}

// PUT /api/v1/orders/:id/eta
func (h *OrderAPI) SetETA(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		EstimatedReadyAt time.Time `json:"estimated_ready_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetOrderETA(c.Request.Context(), id, body.EstimatedReadyAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ord, _ := h.svc.GetOrder(c.Request.Context(), id)
	if h.hub != nil {
		h.hub.Broadcast(struct {
			Type  string        `json:"type"`
			Order *models.Order `json:"order"`
		}{Type: "order_eta_updated", Order: ord})
	}
	c.JSON(http.StatusOK, ord)
}

// GET /api/v1/orders/customer/:customer_id
func (h *OrderAPI) ListOrdersByCustomer(c *gin.Context) {
	cid := c.Param("customer_id")
	res, err := h.svc.ListOrdersByCustomer(c.Request.Context(), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": res})
}

// POST /api/v1/orders/:id/reorder
func (h *OrderAPI) Reorder(c *gin.Context) {
	id := c.Param("id")
	ord, err := h.svc.Reorder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ord)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Retrieve order details by order ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.Order
// @Failure 404 {object} models.ErrorResponse
// @Router /orders/{id} [get]
func (h *OrderAPI) GetOrder(c *gin.Context) {
	id := c.Param("id")
	ord, err := h.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, ord)
}

// ListOrders godoc
// @Summary List all orders
// @Description Get a list of all orders
// @Tags orders
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /orders [get]
func (h *OrderAPI) ListOrders(c *gin.Context) {
	res, err := h.svc.ListOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": res})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param request body models.UpdateOrderStatusRequest true "Status update request"
// @Success 200 {object} models.Order
// @@Failure 400 {object} models.ErrorRespons
// @Router /orders/{id}/status [put]
func (h *OrderAPI) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var body models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateOrderStatus(c.Request.Context(), id, body.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ord, _ := h.svc.GetOrder(c.Request.Context(), id)
	if h.hub != nil {
		h.hub.Broadcast(struct {
			Type  string        `json:"type"`
			Order *models.Order `json:"order"`
		}{Type: "order_updated", Order: ord})
	}
	c.JSON(http.StatusOK, ord)
}
