package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CategoriesAPI struct {
	svc *services.MenuSQLService
}

func NewCategoriesAPI(svc *services.MenuSQLService) *CategoriesAPI { return &CategoriesAPI{svc: svc} }

// CreateCategory godoc
// @Summary Create a category
// @Description Create a new menu category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Category true "Category request"
// @Success 201 {object} models.Category
// @@Failure 400 {object} models.ErrorRespons
// @Router /categories [post]
func (h *CategoriesAPI) CreateCategory(c *gin.Context) {
	var body models.Category
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.svc.CreateCategory(c.Request.Context(), &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// ListCategories godoc
// @Summary List categories
// @Description Get all menu categories
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /categories [get]
func (h *CategoriesAPI) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update menu category details
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param request body models.Category true "Category update"
// @Success 200 {object} models.Category
// @@Failure 400 {object} models.ErrorRespons
// @Router /categories/{id} [put]
func (h *CategoriesAPI) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var body models.Category
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = id
	if err := h.svc.UpdateCategory(c.Request.Context(), &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a menu category
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 "Deleted"
// @@Failure 400 {object} models.ErrorRespons
// @Router /categories/{id} [delete]
func (h *CategoriesAPI) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
