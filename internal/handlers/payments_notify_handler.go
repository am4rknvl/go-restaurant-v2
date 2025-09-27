package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TelebirrNotifyHandler handles Telebirr payment gateway callbacks
func TelebirrNotifyHandler(c *gin.Context) {
	var payload map[string]string
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Resolve a callback-capable service stored in context
	svcVal, ok := c.Get("paymentService")
	if !ok || svcVal == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}
	type telebirrCallback interface{ HandleTelebirrCallback(map[string]string) error }
	svc, ok := svcVal.(telebirrCallback)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service missing callback"})
		return
	}

	if err := svc.HandleTelebirrCallback(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
