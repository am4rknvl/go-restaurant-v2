package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAPI bundles admin-facing placeholder handlers
type AdminAPI struct{}

func NewAdminAPI() *AdminAPI { return &AdminAPI{} }

// Inventory
func (h *AdminAPI) ListInventory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) CreateInventoryItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) AdjustInventory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) LinkRecipe(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Reports
func (h *AdminAPI) SalesReport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) PopularItemsReport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) CustomersReport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) OperationsReport(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Branches
func (h *AdminAPI) ListBranches(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) CreateBranch(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) UpdateBranch(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}

// Staff
func (h *AdminAPI) ListStaff(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) CreateStaff(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
func (h *AdminAPI) UpdateStaff(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented"})
}
