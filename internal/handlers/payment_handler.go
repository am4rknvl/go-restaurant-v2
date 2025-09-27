package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-system/internal/services"
)

type PaymentHandler struct {
	svc services.PaymentService
}

func NewPaymentHandler(svc services.PaymentService) *PaymentHandler { return &PaymentHandler{svc: svc} }

// POST /api/v1/payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	restaurantID := getRestaurantIDFromContext(c)
	var body struct {
		OrderID     uint   `json:"order_id" binding:"required"`
		AmountCents int64  `json:"amount_cents" binding:"required"`
		Provider    string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.CreatePayment(c.Request.Context(), restaurantID, body.OrderID, body.AmountCents, body.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	restaurantID := getRestaurantIDFromContext(c)
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.svc.GetPayment(c.Request.Context(), restaurantID, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "record not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// POST /api/v1/payments/:id/refund
func (h *PaymentHandler) RequestRefund(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Reason string  `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.RequestRefund(c.Request.Context(), id, body.Amount, body.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

// POST /api/v1/payments/partial
func (h *PaymentHandler) ApplyPartialPayment(c *gin.Context) {
	var body struct {
		OrderID string  `json:"order_id" binding:"required"`
		Amount  float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ApplyPartialPayment(c.Request.Context(), body.OrderID, body.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
