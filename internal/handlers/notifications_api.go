package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type NotificationsAPI struct {
	svc *services.NotificationService
}

func NewNotificationsAPI(svc *services.NotificationService) *NotificationsAPI {
	return &NotificationsAPI{svc: svc}
}

// Subscribe godoc
// @Summary Subscribe to notifications
// @Description Subscribe to push notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Subscription true "Subscription request"
// @Success 201 {object} models.Subscription
// @@Failure 400 {object} models.ErrorRespons
// @Router /notifications/subscribe [post]
func (h *NotificationsAPI) Subscribe(c *gin.Context) {
	var s models.Subscription
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Subscribe(c.Request.Context(), &s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// Unsubscribe godoc
// @Summary Unsubscribe from notifications
// @Description Remove notification subscription
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Success 200 "Unsubscribed"
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/{id} [delete]
func (h *NotificationsAPI) Unsubscribe(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Unsubscribe(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// ListForAccount godoc
// @Summary List notifications for account
// @Description Get all notification subscriptions for an account
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param account_id path string true "Account ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/account/{account_id} [get]
func (h *NotificationsAPI) ListForAccount(c *gin.Context) {
	aid := c.Param("account_id")
	res, err := h.svc.ListForAccount(c.Request.Context(), aid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": res})
}

// SendNotification godoc
// @Summary Send notification
// @Description Send a push notification (staff/admin)
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{subscription=models.Subscription,message=string} true "Notification payload"
// @Success 200 "Notification sent"
// @@Failure 400 {object} models.ErrorRespons
// @Router /kitchen/notifications/send [post]
func (h *NotificationsAPI) SendNotification(c *gin.Context) {
	var payload struct {
		Subscription models.Subscription `json:"subscription" binding:"required"`
		Message      string              `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SendNotification(c.Request.Context(), payload.Subscription, map[string]string{"message": payload.Message}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
