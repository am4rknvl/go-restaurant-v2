package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type TelebirrB2BHandler struct {
	telebirrService *services.TelebirrService
	orderService    interface {
		GetOrder(orderID string) (*models.Order, error)
		UpdateOrderStatus(orderID, status string) error
	}
}

func NewTelebirrB2BHandler(telebirrService *services.TelebirrService, orderService interface {
	GetOrder(orderID string) (*models.Order, error)
	UpdateOrderStatus(orderID, status string) error
}) *TelebirrB2BHandler {
	return &TelebirrB2BHandler{
		telebirrService: telebirrService,
		orderService:    orderService,
	}
}

type CreatePaymentRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Subject string  `json:"subject" binding:"required"`
	Body    string  `json:"body"`
}

type CreatePaymentResponse struct {
	PrepayID   string `json:"prepay_id"`
	PaymentURL string `json:"payment_url"`
	Status     string `json:"status"`
}

// CreateB2BPayment godoc
// @Summary Create Telebirr B2B payment
// @Description Create a new Telebirr B2B payment for an order
// @Tags telebirr-b2b
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePaymentRequest true "Payment request"
// @Success 201 {object} CreatePaymentResponse
// @@Failure 400 {object} models.ErrorRespons
// @Failure 404 {object} models.ErrorResponse
// @Router /payments/telebirr/b2b/create [post]
func (h *TelebirrB2BHandler) CreateB2BPayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate order exists
	_, err := h.orderService.GetOrder(req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order_not_found"})
		return
	}

	// Create prepaid order with Telebirr
	telebirrOrder, err := h.telebirrService.CreatePrepaidOrder(
		req.OrderID,
		req.Amount,
		req.Subject,
		req.Body,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_prepaid_order", "details": err.Error()})
		return
	}

	// Generate payment URL
	paymentURL, err := h.telebirrService.GeneratePaymentURL(telebirrOrder.PrepayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_generate_payment_url", "details": err.Error()})
		return
	}

	response := CreatePaymentResponse{
		PrepayID:   telebirrOrder.PrepayID,
		PaymentURL: paymentURL,
		Status:     telebirrOrder.Status,
	}

	c.JSON(http.StatusCreated, response)
}

// GetPaymentStatus godoc
// @Summary Get Telebirr payment status
// @Description Get the status of a Telebirr B2B payment
// @Tags telebirr-b2b
// @Produce json
// @Param prepay_id path string true "Prepay ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /payments/telebirr/b2b/status/{prepay_id} [get]
func (h *TelebirrB2BHandler) GetPaymentStatus(c *gin.Context) {
	prepayID := c.Param("prepay_id")
	if prepayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prepay_id_required"})
		return
	}

	var telebirrOrder models.TelebirrOrder
	if err := h.telebirrService.DB().Where("prepay_id = ?", prepayID).First(&telebirrOrder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prepay_id":      telebirrOrder.PrepayID,
		"merch_order_id": telebirrOrder.MerchOrderID,
		"order_id":       telebirrOrder.OrderID,
		"amount":         telebirrOrder.Amount,
		"status":         telebirrOrder.Status,
		"payment_url":    telebirrOrder.PaymentURL,
		"created_at":     telebirrOrder.CreatedAt,
		"updated_at":     telebirrOrder.UpdatedAt,
	})
}

// HandleB2BNotification godoc
// @Summary Handle Telebirr payment notification
// @Description Webhook endpoint for Telebirr payment notifications
// @Tags telebirr-b2b
// @Accept x-www-form-urlencoded
// @Produce plain
// @Success 200 {string} string "success"
// @@Failure 400 {object} models.ErrorRespons
// @Router /payments/telebirr/b2b/notify [post]
func (h *TelebirrB2BHandler) HandleB2BNotification(c *gin.Context) {
	// Parse form data from Telebirr notification
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_form_data"})
		return
	}

	notification := make(map[string]string)
	for key, values := range c.Request.Form {
		if len(values) > 0 {
			notification[key] = values[0]
		}
	}

	// Process the notification
	if err := h.telebirrService.ProcessNotification(notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification_processing_failed", "details": err.Error()})
		return
	}

	// Update order status based on trade status
	prepayID := notification["prepay_id"]
	tradeStatus := notification["trade_status"]

	var telebirrOrder models.TelebirrOrder
	if err := h.telebirrService.DB().Where("prepay_id = ?", prepayID).First(&telebirrOrder).Error; err == nil {
		var orderStatus string
		switch tradeStatus {
		case "TRADE_SUCCESS":
			orderStatus = "paid"
		case "TRADE_CLOSED":
			orderStatus = "cancelled"
		case "WAIT_BUYER_PAY":
			orderStatus = "pending_payment"
		default:
			orderStatus = "pending_payment"
		}

		if err := h.orderService.UpdateOrderStatus(telebirrOrder.OrderID, orderStatus); err != nil {
			// Log error but don't fail the notification response
			// Telebirr expects a success response to avoid retries
		}
	}

	// Telebirr expects "success" response
	c.String(http.StatusOK, "success")
}

// HandleB2BReturn godoc
// @Summary Handle Telebirr payment return
// @Description Handle user return from Telebirr payment page
// @Tags telebirr-b2b
// @Produce json
// @Param prepay_id query string true "Prepay ID"
// @Param trade_status query string false "Trade status"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /payments/telebirr/b2b/return [get]
func (h *TelebirrB2BHandler) HandleB2BReturn(c *gin.Context) {
	prepayID := c.Query("prepay_id")
	tradeStatus := c.Query("trade_status")

	if prepayID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_prepay_id"})
		return
	}

	var telebirrOrder models.TelebirrOrder
	if err := h.telebirrService.DB().Where("prepay_id = ?", prepayID).First(&telebirrOrder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment_not_found"})
		return
	}

	// Determine redirect based on trade status
	var redirectURL string
	var message string

	switch tradeStatus {
	case "TRADE_SUCCESS":
		redirectURL = "/payment/success"
		message = "Payment completed successfully"
	case "TRADE_CLOSED":
		redirectURL = "/payment/failed"
		message = "Payment was cancelled or failed"
	default:
		redirectURL = "/payment/pending"
		message = "Payment is being processed"
	}

	// Return JSON response for API clients or redirect for web clients
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, gin.H{
			"prepay_id":    prepayID,
			"trade_status": tradeStatus,
			"order_id":     telebirrOrder.OrderID,
			"message":      message,
			"redirect_url": redirectURL,
		})
	} else {
		// Redirect for web browsers
		c.Redirect(http.StatusFound, redirectURL+"?order_id="+telebirrOrder.OrderID+"&status="+tradeStatus)
	}
}

// GetOrderPayments godoc
// @Summary Get order payments
// @Description Get all Telebirr payments for an order
// @Tags telebirr-b2b
// @Produce json
// @Security BearerAuth
// @Param order_id path string true "Order ID"
// @Success 200 {object} map[string]interface{}
// @@Failure 400 {object} models.ErrorRespons
// @Router /payments/telebirr/b2b/orders/{order_id} [get]
func (h *TelebirrB2BHandler) GetOrderPayments(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id_required"})
		return
	}

	var telebirrOrders []models.TelebirrOrder
	if err := h.telebirrService.DB().Where("order_id = ?", orderID).Find(&telebirrOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"payments": telebirrOrders,
	})
}

// RefundB2BPayment godoc
// @Summary Refund Telebirr payment
// @Description Process a refund for a Telebirr B2B payment
// @Tags telebirr-b2b
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{prepay_id=string,refund_amount=number,refund_reason=string} true "Refund request"
// @Success 200 {object} map[string]interface{}
// @@Failure 400 {object} models.ErrorRespons
// @Router /payments/telebirr/b2b/refund [post]
func (h *TelebirrB2BHandler) RefundB2BPayment(c *gin.Context) {
	var req struct {
		PrepayID     string  `json:"prepay_id" binding:"required"`
		RefundAmount float64 `json:"refund_amount" binding:"required"`
		RefundReason string  `json:"refund_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var telebirrOrder models.TelebirrOrder
	if err := h.telebirrService.DB().Where("prepay_id = ? AND status = 'completed'", req.PrepayID).First(&telebirrOrder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "completed_payment_not_found"})
		return
	}

	if req.RefundAmount > telebirrOrder.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refund_amount_exceeds_payment_amount"})
		return
	}

	// TODO: Implement actual refund API call to Telebirr
	// For now, just update local status
	telebirrOrder.Status = "refunded"
	if err := h.telebirrService.DB().Save(&telebirrOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_payment_status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prepay_id":     req.PrepayID,
		"refund_amount": req.RefundAmount,
		"status":        "refunded",
		"message":       "Refund processed successfully",
	})
}
