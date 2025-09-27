package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MenuAdminAPI struct {
	svc *services.MenuSQLService
}

func NewMenuAdminAPI(svc *services.MenuSQLService) *MenuAdminAPI { return &MenuAdminAPI{svc: svc} }

// POST /api/v1/menu/item
func (h *MenuAdminAPI) CreateItem(c *gin.Context) {
	var body models.MenuItem
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.svc.CreateItem(c.Request.Context(), &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// PUT /api/v1/menu/item/:id
func (h *MenuAdminAPI) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var body models.MenuItem
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = id
	if err := h.svc.UpdateItem(c.Request.Context(), &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// DELETE /api/v1/menu/item/:id
func (h *MenuAdminAPI) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteItem(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
