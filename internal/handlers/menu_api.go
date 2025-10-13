package handlers

import (
	"restaurant-system/internal/errors"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type MenuAPI struct {
	svc *services.MenuSQLService
}

func NewMenuAPI(svc *services.MenuSQLService) *MenuAPI { return &MenuAPI{svc: svc} }

// GetQRMenu godoc
// @Summary Get QR menu
// @Description Get menu for a specific restaurant table via QR code
// @Tags menu
// @Produce json
// @Param restaurant_id path string true "Restaurant ID"
// @Param table_id path string true "Table ID"
// @Param lang query string false "Language code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /restaurant/{restaurant_id}/table/{table_id}/menu [get]
func (h *MenuAPI) GetQRMenu(c *gin.Context) {
	restaurantID := c.Param("restaurant_id")
	tableID := c.Param("table_id")
	lang := c.Query("lang")

	// Validate required parameters
	if restaurantID == "" {
		errors.RespondError(c, errors.NewValidation("restaurant_id is required"))
		return
	}
	if tableID == "" {
		errors.RespondError(c, errors.NewValidation("table_id is required"))
		return
	}

	cats, err := h.svc.GetQRMenu(c.Request.Context(), restaurantID, tableID, lang)
	if err != nil {
		errors.HandleDatabaseError(c, err, "retrieve menu")
		return
	}

	errors.RespondOK(c, gin.H{"categories": cats})
}
