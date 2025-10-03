package handlers

import (
	"net/http"
	"strings"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type TelebirrC2BHandler struct {
	c2bService   *services.TelebirrC2BService
	orderService interface {
		GetOrder(orderID string) (*models.Order, error)
		UpdateOrderStatus(orderID, status string) error
	}
}

func NewTelebirrC2BHandler(c2bService *services.TelebirrC2BService, orderService interface {
	GetOrder(orderID string) (*models.Order, error)
	UpdateOrderStatus(orderID, status string) error
}) *TelebirrC2BHandler {
	return &TelebirrC2BHandler{
		c2bService:   c2bService,
		orderService: orderService,
	}
}

type CreateC2BPaymentRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Subject string  `json:"subject" binding:"required"`
	Body    string  `json:"body"`
}

type CreateC2BPaymentResponse struct {
	OutTradeNo string `json:"out_trade_no"`
	H5PayURL   string `json:"h5_pay_url"`
	TradeNo    string `json:"trade_no"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// POST /api/v1/payments/telebirr/c2b/create
func (h *TelebirrC2BHandler) CreateC2BPayment(c *gin.Context) {
	var req CreateC2BPaymentRequest
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

	// Validate amount matches order (optional check)
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_amount"})
		return
	}

	// Create H5 C2B payment
	c2bOrder, err := h.c2bService.CreateH5Payment(
		req.OrderID,
		req.Amount,
		req.Subject,
		req.Body,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_h5_payment", "details": err.Error()})
		return
	}

	response := CreateC2BPaymentResponse{
		OutTradeNo: c2bOrder.OutTradeNo,
		H5PayURL:   c2bOrder.H5PayURL,
		TradeNo:    c2bOrder.TradeNo,
		Status:     c2bOrder.Status,
		Message:    "H5 payment created successfully. Redirect customer to h5_pay_url",
	}

	c.JSON(http.StatusCreated, response)
}

// GET /api/v1/payments/telebirr/c2b/status/:out_trade_no
func (h *TelebirrC2BHandler) GetC2BPaymentStatus(c *gin.Context) {
	outTradeNo := c.Param("out_trade_no")
	if outTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "out_trade_no_required"})
		return
	}

	c2bOrder, err := h.c2bService.GetOrderByOutTradeNo(outTradeNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"out_trade_no": c2bOrder.OutTradeNo,
		"order_id":     c2bOrder.OrderID,
		"trade_no":     c2bOrder.TradeNo,
		"total_amount": c2bOrder.TotalAmount,
		"status":       c2bOrder.Status,
		"h5_pay_url":   c2bOrder.H5PayURL,
		"subject":      c2bOrder.Subject,
		"created_at":   c2bOrder.CreatedAt,
		"updated_at":   c2bOrder.UpdatedAt,
	})
}

// POST /api/v1/payments/telebirr/c2b/notify
func (h *TelebirrC2BHandler) HandleC2BNotification(c *gin.Context) {
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

	// Process the C2B notification
	if err := h.c2bService.ProcessC2BNotification(notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification_processing_failed", "details": err.Error()})
		return
	}

	// Update restaurant order status based on trade status
	tradeStatus := notification["trade_status"]

	// Extract order_id from passback_params
	orderID := h.extractOrderIDFromPassback(notification["passback_params"])
	if orderID != "" {
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

		if err := h.orderService.UpdateOrderStatus(orderID, orderStatus); err != nil {
			// Log error but don't fail the notification response
			// Telebirr expects a success response to avoid retries
		}
	}

	// Telebirr expects "success" response
	c.String(http.StatusOK, "success")
}

// GET /api/v1/payments/telebirr/c2b/return
func (h *TelebirrC2BHandler) HandleC2BReturn(c *gin.Context) {
	outTradeNo := c.Query("out_trade_no")
	tradeStatus := c.Query("trade_status")

	if outTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_out_trade_no"})
		return
	}

	c2bOrder, err := h.c2bService.GetOrderByOutTradeNo(outTradeNo)
	if err != nil {
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
			"out_trade_no": outTradeNo,
			"trade_status": tradeStatus,
			"order_id":     c2bOrder.OrderID,
			"message":      message,
			"redirect_url": redirectURL,
		})
	} else {
		// Redirect for web browsers
		c.Redirect(http.StatusFound, redirectURL+"?order_id="+c2bOrder.OrderID+"&status="+tradeStatus)
	}
}

// GET /api/v1/payments/telebirr/c2b/orders/:order_id
func (h *TelebirrC2BHandler) GetOrderC2BPayments(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id_required"})
		return
	}

	c2bOrders, err := h.c2bService.GetOrdersByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"payments": c2bOrders,
		"count":    len(c2bOrders),
	})
}

// POST /api/v1/payments/telebirr/c2b/refund
func (h *TelebirrC2BHandler) RefundC2BPayment(c *gin.Context) {
	var req struct {
		OutTradeNo   string  `json:"out_trade_no" binding:"required"`
		RefundAmount float64 `json:"refund_amount" binding:"required"`
		RefundReason string  `json:"refund_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c2bOrder, err := h.c2bService.GetOrderByOutTradeNo(req.OutTradeNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment_not_found"})
		return
	}

	if c2bOrder.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only_completed_payments_can_be_refunded"})
		return
	}

	if req.RefundAmount > c2bOrder.TotalAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refund_amount_exceeds_payment_amount"})
		return
	}

	if err := h.c2bService.RefundC2B(req.OutTradeNo, req.RefundAmount, req.RefundReason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refund_failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"out_trade_no":  req.OutTradeNo,
		"refund_amount": req.RefundAmount,
		"status":        "refunded",
		"message":       "Refund processed successfully",
	})
}

// Helper function to extract order_id from passback_params
func (h *TelebirrC2BHandler) extractOrderIDFromPassback(passbackParams string) string {
	if passbackParams == "" {
		return ""
	}

	// Parse passback_params (format: "order_id=order123&other=value")
	params := make(map[string]string)
	pairs := strings.Split(passbackParams, "&")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}

	return params["order_id"]
}

// GET /api/v1/payments/telebirr/c2b/query/:out_trade_no
func (h *TelebirrC2BHandler) QueryC2BPayment(c *gin.Context) {
	outTradeNo := c.Param("out_trade_no")
	if outTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "out_trade_no_required"})
		return
	}

	// Call Telebirr query API for real-time status
	res, err := h.c2bService.QueryC2B(outTradeNo)
	if err != nil {
		// Fallback to local DB if query fails
		c2bOrder, dbErr := h.c2bService.GetOrderByOutTradeNo(outTradeNo)
		if dbErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query_failed", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"out_trade_no": c2bOrder.OutTradeNo,
			"trade_no":     c2bOrder.TradeNo,
			"trade_status": h.mapStatusToTradeStatus(c2bOrder.Status),
			"total_amount": c2bOrder.TotalAmount,
			"subject":      c2bOrder.Subject,
			"gmt_create":   c2bOrder.CreatedAt.Format("2006-01-02 15:04:05"),
			"gmt_payment":  c2bOrder.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"out_trade_no": outTradeNo,
		"trade_no":     res.TradeNo,
		"trade_status": res.TradeStatus,
		"total_amount": res.TotalAmount,
		"subject":      res.Subject,
		"gmt_create":   res.GmtCreate,
		"gmt_payment":  res.GmtPayment,
	})
}

func (h *TelebirrC2BHandler) mapStatusToTradeStatus(status string) string {
	switch status {
	case "completed":
		return "TRADE_SUCCESS"
	case "failed":
		return "TRADE_CLOSED"
	case "pending":
		return "WAIT_BUYER_PAY"
	case "refunded":
		return "TRADE_REFUNDED"
	default:
		return "WAIT_BUYER_PAY"
	}
}
