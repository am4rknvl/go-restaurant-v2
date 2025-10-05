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

// CreatePayment godoc
// @Summary Create a payment
// @Description Create a new payment for an order
// @Tags payments
// @Accept json
// @Produce json
// @Param request body object{order_id=int,amount_cents=int,provider=string} true "Payment request"
// @Success 201 {object} models.Payment
// @@Failure 400 {object} models.ErrorRespons
// @Router /payments [post]
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

// GetPayment godoc
// @Summary Get payment by ID
// @Description Retrieve payment details
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} models.Payment
// @Failure 404 {object} models.ErrorResponse
// @Router /payments/{id} [get]
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

// RequestRefund godoc
// @Summary Request payment refund
// @Description Request a refund for a payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body object{amount=number,reason=string} true "Refund request"
// @Success 201 {object} models.Refund
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/{id}/refund [post]
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

// ApplyPartialPayment godoc
// @Summary Apply partial payment
// @Description Apply a partial payment to an order
// @Tags payments
// @Accept json
// @Produce json
// @Param request body object{order_id=string,amount=number} true "Partial payment"
// @Success 200 "Payment applied"
// @Failure 500 {object} models.ErrorResponse
// @Router /payments/partial [post]
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
