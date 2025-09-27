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

// POST /api/v1/notifications/subscribe
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

// DELETE /api/v1/notifications/:id
func (h *NotificationsAPI) Unsubscribe(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Unsubscribe(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// GET /api/v1/notifications/account/:account_id
func (h *NotificationsAPI) ListForAccount(c *gin.Context) {
	aid := c.Param("account_id")
	res, err := h.svc.ListForAccount(c.Request.Context(), aid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": res})
}

// POST /api/v1/notifications/send (staff/admin trigger)
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
