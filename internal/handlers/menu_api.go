package handlers

import (
	"net/http"

	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type MenuAPI struct {
	svc *services.MenuSQLService
}

func NewMenuAPI(svc *services.MenuSQLService) *MenuAPI { return &MenuAPI{svc: svc} }

// GET /api/v1/restaurant/:restaurant_id/table/:table_id/menu
func (h *MenuAPI) GetQRMenu(c *gin.Context) {
	restaurantID := c.Param("restaurant_id")
	tableID := c.Param("table_id")
	lang := c.Query("lang")
	cats, err := h.svc.GetQRMenu(c.Request.Context(), restaurantID, tableID, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}
