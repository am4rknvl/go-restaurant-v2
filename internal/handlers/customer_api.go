package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomerAPI bundles customer-facing placeholder handlers
type CustomerAPI struct{}

func NewCustomerAPI() *CustomerAPI { return &CustomerAPI{} }

// Profiles
func (h *CustomerAPI) Me(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *CustomerAPI) UpdateMe(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Loyalty & Promo
func (h *CustomerAPI) GetLoyalty(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *CustomerAPI) RedeemPoints(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *CustomerAPI) ApplyPromo(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Waitlist
func (h *CustomerAPI) CreateWaitlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *CustomerAPI) UpdateWaitlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
