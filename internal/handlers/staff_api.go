package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StaffAPI bundles staff-facing placeholder handlers
type StaffAPI struct{}

func NewStaffAPI() *StaffAPI { return &StaffAPI{} }

// UpdateTableState godoc
// @Summary Update table state
// @Description Update the state of a table (staff)
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Param tableId path string true "Table ID"
// @Success 501 {object} map[string]string
// @Router /staff/branches/{branchId}/tables/{tableId}/state [patch]
func (h *StaffAPI) UpdateTableState(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// AssignWaiterToTable godoc
// @Summary Assign waiter to table
// @Description Assign a waiter to a specific table
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Param tableId path string true "Table ID"
// @Success 501 {object} map[string]string
// @Router /staff/branches/{branchId}/tables/{tableId}/assign [post]
func (h *StaffAPI) AssignWaiterToTable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// AssignChefToOrder godoc
// @Summary Assign chef to order
// @Description Assign a chef to prepare an order
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Param orderId path string true "Order ID"
// @Success 501 {object} map[string]string
// @Router /staff/branches/{branchId}/orders/{orderId}/assign-chef [post]
func (h *StaffAPI) AssignChefToOrder(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// SplitOrder godoc
// @Summary Split an order
// @Description Split an order into multiple orders
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Param orderId path string true "Order ID"
// @Success 501 {object} map[string]string
// @Router /staff/branches/{branchId}/orders/{orderId}/split [post]
func (h *StaffAPI) SplitOrder(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// MergeOrders godoc
// @Summary Merge orders
// @Description Merge multiple orders into one
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Success 501 {object} map[string]string
// @Router /staff/branches/{branchId}/orders/merge [post]
func (h *StaffAPI) MergeOrders(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// AddTip godoc
// @Summary Add tip to order
// @Description Add a tip amount to an order
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param branchId path string true "Branch ID"
// @Param orderId path string true "Order ID"
// @Success 501 {object} map[string]string
// @Router /staff/branches/{branchId}/orders/{orderId}/tip [post]
func (h *StaffAPI) AddTip(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
