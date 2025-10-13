package handlers

import (
	"restaurant-system/internal/errors"
	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MenuAdminAPI struct {
	svc *services.MenuSQLService
}

func NewMenuAdminAPI(svc *services.MenuSQLService) *MenuAdminAPI { return &MenuAdminAPI{svc: svc} }

// CreateItem godoc
// @Summary Create menu item (admin)
// @Description Create a new menu item (admin route)
// @Tags menu-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.MenuItem true "Menu item request"
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} models.ErrorResponse
// @Router /menu/item [post]
func (h *MenuAdminAPI) CreateItem(c *gin.Context) {
	var body models.MenuItem
	if err := c.ShouldBindJSON(&body); err != nil {
		errors.HandleBindError(c, err)
		return
	}

	// Validate required fields
	if body.Name == "" {
		errors.RespondError(c, errors.NewValidation("name is required"))
		return
	}
	if body.Category == "" {
		errors.RespondError(c, errors.NewValidation("category is required"))
		return
	}
	if body.Price <= 0 {
		errors.RespondError(c, errors.NewValidation("price must be greater than 0"))
		return
	}

	if body.ID == "" {
		body.ID = uuid.New().String()
	}

	if err := h.svc.CreateItem(c.Request.Context(), &body); err != nil {
		errors.HandleDatabaseError(c, err, "create menu item")
		return
	}

	errors.RespondCreated(c, body)
}

// UpdateItem godoc
// @Summary Update menu item (admin)
// @Description Update menu item details (admin route)
// @Tags menu-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Item ID"
// @Param request body models.MenuItem true "Menu item update"
// @Success 200 {object} models.MenuItem
// @Failure 400 {object} models.ErrorResponse
// @Router /menu/item/{id} [put]
func (h *MenuAdminAPI) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.RespondError(c, errors.NewValidation("id is required"))
		return
	}

	var body models.MenuItem
	if err := c.ShouldBindJSON(&body); err != nil {
		errors.HandleBindError(c, err)
		return
	}

	body.ID = id
	if err := h.svc.UpdateItem(c.Request.Context(), &body); err != nil {
		errors.HandleDatabaseError(c, err, "update menu item")
		return
	}

	errors.RespondOK(c, body)
}

// DeleteItem godoc
// @Summary Delete menu item (admin)
// @Description Delete a menu item (admin route)
// @Tags menu-admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Menu Item ID"
// @Success 204 "Deleted"
// @Failure 400 {object} models.ErrorResponse
// @Router /menu/item/{id} [delete]
func (h *MenuAdminAPI) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.RespondError(c, errors.NewValidation("id is required"))
		return
	}

	if err := h.svc.DeleteItem(c.Request.Context(), id); err != nil {
		errors.HandleDatabaseError(c, err, "delete menu item")
		return
	}

	errors.RespondNoContent(c)
}
