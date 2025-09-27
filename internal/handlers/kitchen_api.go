package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type KitchenAPI struct {
	svc *services.OrderSQLService
}

func NewKitchenAPI(svc *services.OrderSQLService) *KitchenAPI { return &KitchenAPI{svc: svc} }

// GET /api/v1/kitchen/orders
func (h *KitchenAPI) ListPending(c *gin.Context) {
	res, err := h.svc.ListOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// filter pending-like statuses in handler for simplicity
	var pending []*models.Order
	for _, o := range res {
		if o.Status == models.OrderStatusPending || o.Status == models.OrderStatusConfirmed || o.Status == models.OrderStatusPreparing {
			pending = append(pending, o)
		}
	}
	c.JSON(http.StatusOK, gin.H{"orders": pending})
}

// PUT /api/v1/kitchen/orders/:id/status
func (h *KitchenAPI) UpdateStatus(c *gin.Context) {
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
	c.JSON(http.StatusOK, ord)
}
