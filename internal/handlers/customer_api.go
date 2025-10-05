package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomerAPI bundles customer-facing placeholder handlers
type CustomerAPI struct{}

func NewCustomerAPI() *CustomerAPI { return &CustomerAPI{} }

// Me godoc
// @Summary Get current user profile
// @Description Get authenticated user's profile
// @Tags customer
// @Produce json
// @Security BearerAuth
// @Success 501 {object} map[string]string
// @Router /me [get]
func (h *CustomerAPI) Me(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// UpdateMe godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 501 {object} map[string]string
// @Router /me [patch]
func (h *CustomerAPI) UpdateMe(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// GetLoyalty godoc
// @Summary Get loyalty points
// @Description Get user's loyalty points balance
// @Tags customer
// @Produce json
// @Security BearerAuth
// @Success 501 {object} map[string]string
// @Router /loyalty [get]
func (h *CustomerAPI) GetLoyalty(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// RedeemPoints godoc
// @Summary Redeem loyalty points
// @Description Redeem loyalty points for rewards
// @Tags customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 501 {object} map[string]string
// @Router /loyalty/redeem [post]
func (h *CustomerAPI) RedeemPoints(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// ApplyPromo godoc
// @Summary Apply promo code
// @Description Apply a promotional code to order
// @Tags customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 501 {object} map[string]string
// @Router /promo/apply [post]
func (h *CustomerAPI) ApplyPromo(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// CreateWaitlist godoc
// @Summary Join waitlist
// @Description Add customer to restaurant waitlist
// @Tags customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Success 501 {object} map[string]string
// @Router /branches/{branchId}/waitlist [post]
func (h *CustomerAPI) CreateWaitlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// UpdateWaitlist godoc
// @Summary Update waitlist entry
// @Description Update customer's waitlist entry
// @Tags customer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Param entryId path string true "Entry ID"
// @Success 501 {object} map[string]string
// @Router /branches/{branchId}/waitlist/{entryId} [patch]
func (h *CustomerAPI) UpdateWaitlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
