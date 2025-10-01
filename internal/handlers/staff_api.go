package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StaffAPI bundles staff-facing placeholder handlers
type StaffAPI struct{}

func NewStaffAPI() *StaffAPI { return &StaffAPI{} }

// Table states
func (h *StaffAPI) UpdateTableState(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Assignments
func (h *StaffAPI) AssignWaiterToTable(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *StaffAPI) AssignChefToOrder(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Order lifecycle extensions
func (h *StaffAPI) SplitOrder(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *StaffAPI) MergeOrders(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *StaffAPI) AddTip(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
