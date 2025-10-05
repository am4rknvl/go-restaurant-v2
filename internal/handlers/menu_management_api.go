package handlers

import (
	"net/http"

	"restaurant-system/internal/auth"
	"restaurant-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MenuManagementAPI struct {
	gorm *gorm.DB
	ws   interface{ Broadcast(v interface{}) }
}

func NewMenuManagementAPI(gdb *gorm.DB, ws interface{ Broadcast(v interface{}) }) *MenuManagementAPI {
	return &MenuManagementAPI{gorm: gdb, ws: ws}
}

// CreateCategory godoc
// @Summary Create menu category
// @Description Create a new menu category
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.MenuCategory true "Category request"
// @Success 201 {object} models.MenuCategory
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/categories [post]
func (h *MenuManagementAPI) CreateCategory(c *gin.Context) {
	var body models.MenuCategory
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.gorm.Create(&body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Retrieve menu category details
// @Tags menu
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.MenuCategory
// @Failure 404 {object} models.ErrorResponse
// @Router /menu/categories/{id} [get]
func (h *MenuManagementAPI) GetCategory(c *gin.Context) {
	var cat models.MenuCategory
	if err := h.gorm.First(&cat, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// ListCategories godoc
// @Summary List menu categories
// @Description Get all menu categories
// @Tags menu
// @Produce json
// @Param restaurant_id query string false "Filter by restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /menu/categories [get]
func (h *MenuManagementAPI) ListCategories(c *gin.Context) {
	var cats []models.MenuCategory
	q := h.gorm.Model(&models.MenuCategory{})
	if rid := c.Query("restaurant_id"); rid != "" {
		q = q.Where("restaurant_id = ?", rid)
	}
	if err := q.Order("name asc").Find(&cats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

// UpdateCategory godoc
// @Summary Update menu category
// @Description Update category details
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param request body models.MenuCategory true "Category update"
// @Success 200 {object} models.MenuCategory
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/categories/{id} [put]
func (h *MenuManagementAPI) UpdateCategory(c *gin.Context) {
	var body models.MenuCategory
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = c.Param("id")
	if err := h.gorm.Model(&models.MenuCategory{ID: body.ID}).Updates(body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// DeleteCategory godoc
// @Summary Delete menu category
// @Description Delete a category
// @Tags menu
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 "Deleted"
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/categories/{id} [delete]
func (h *MenuManagementAPI) DeleteCategory(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuCategory{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateItem godoc
// @Summary Create menu item
// @Description Create a new menu item
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.MenuItemGorm true "Menu item request"
// @Success 201 {object} models.MenuItemGorm
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/items [post]
func (h *MenuManagementAPI) CreateItem(c *gin.Context) {
	var body models.MenuItemGorm
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.gorm.Create(&body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// GetItem godoc
// @Summary Get menu item by ID
// @Description Retrieve menu item details
// @Tags menu
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} models.MenuItemGorm
// @Failure 404 {object} models.ErrorResponse
// @Router /menu/items/{id} [get]
func (h *MenuManagementAPI) GetItem(c *gin.Context) {
	var it models.MenuItemGorm
	if err := h.gorm.First(&it, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, it)
}

// ListItems godoc
// @Summary List menu items
// @Description Get all menu items
// @Tags menu
// @Produce json
// @Param restaurant_id query string false "Filter by restaurant ID"
// @Param category_id query string false "Filter by category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /menu/items [get]
func (h *MenuManagementAPI) ListItems(c *gin.Context) {
	var items []models.MenuItemGorm
	q := h.gorm.Model(&models.MenuItemGorm{})
	if rid := c.Query("restaurant_id"); rid != "" {
		q = q.Where("restaurant_id = ?", rid)
	}
	if cid := c.Query("category_id"); cid != "" {
		q = q.Where("category_id = ?", cid)
	}
	if err := q.Order("name asc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// UpdateItem godoc
// @Summary Update menu item
// @Description Update menu item details
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body models.MenuItemGorm true "Item update"
// @Success 200 {object} models.MenuItemGorm
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/items/{id} [put]
func (h *MenuManagementAPI) UpdateItem(c *gin.Context) {
	var body models.MenuItemGorm
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = c.Param("id")
	if err := h.gorm.Model(&models.MenuItemGorm{ID: body.ID}).Updates(body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// UpdateAvailability godoc
// @Summary Update item availability
// @Description Update menu item availability status
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body object{available=bool} true "Availability update"
// @Success 204 "Updated"
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/items/{id}/availability [patch]
func (h *MenuManagementAPI) UpdateAvailability(c *gin.Context) {
	type req struct {
		Available bool `json:"available" binding:"required"`
	}
	var body req
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	if err := h.gorm.Model(&models.MenuItemGorm{}).Where("id = ?", id).Update("available", body.Available).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// broadcast availability change
	h.ws.Broadcast(gin.H{"type": "menu.availability", "item_id": id, "available": body.Available})
	c.Status(http.StatusNoContent)
}

// DeleteItem godoc
// @Summary Delete menu item
// @Description Delete a menu item
// @Tags menu
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Success 204 "Deleted"
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/items/{id} [delete]
func (h *MenuManagementAPI) DeleteItem(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuItemGorm{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateVariant godoc
// @Summary Create menu item variant
// @Description Create a variant for a menu item
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body models.MenuVariant true "Variant request"
// @Success 201 {object} models.MenuVariant
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/items/{id}/variants [post]
func (h *MenuManagementAPI) CreateVariant(c *gin.Context) {
	var body models.MenuVariant
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ItemID = c.Param("id")
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.gorm.Create(&body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// UpdateVariant godoc
// @Summary Update menu variant
// @Description Update variant details
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Variant ID"
// @Param request body models.MenuVariant true "Variant update"
// @Success 200 {object} models.MenuVariant
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/variants/{id} [put]
func (h *MenuManagementAPI) UpdateVariant(c *gin.Context) {
	var body models.MenuVariant
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = c.Param("id")
	if err := h.gorm.Model(&models.MenuVariant{ID: body.ID}).Updates(body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// DeleteVariant godoc
// @Summary Delete menu variant
// @Description Delete a variant
// @Tags menu
// @Produce json
// @Security BearerAuth
// @Param id path string true "Variant ID"
// @Success 204 "Deleted"
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/variants/{id} [delete]
func (h *MenuManagementAPI) DeleteVariant(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuVariant{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateAddon godoc
// @Summary Create menu item addon
// @Description Create an addon for a menu item
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body models.MenuAddon true "Addon request"
// @Success 201 {object} models.MenuAddon
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/items/{id}/addons [post]
func (h *MenuManagementAPI) CreateAddon(c *gin.Context) {
	var body models.MenuAddon
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ItemID = c.Param("id")
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.gorm.Create(&body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// UpdateAddon godoc
// @Summary Update menu addon
// @Description Update addon details
// @Tags menu
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Addon ID"
// @Param request body models.MenuAddon true "Addon update"
// @Success 200 {object} models.MenuAddon
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/addons/{id} [put]
func (h *MenuManagementAPI) UpdateAddon(c *gin.Context) {
	var body models.MenuAddon
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = c.Param("id")
	if err := h.gorm.Model(&models.MenuAddon{ID: body.ID}).Updates(body).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

// DeleteAddon godoc
// @Summary Delete menu addon
// @Description Delete an addon
// @Tags menu
// @Produce json
// @Security BearerAuth
// @Param id path string true "Addon ID"
// @Success 204 "Deleted"
// @@Failure 400 {object} models.ErrorRespons
// @Router /menu/addons/{id} [delete]
func (h *MenuManagementAPI) DeleteAddon(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuAddon{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Role middleware wrappers
func RequireAdminOrManager() gin.HandlerFunc { return auth.RequireAnyRole("admin", "manager") }
func RequireStaff() gin.HandlerFunc {
	return auth.RequireAnyRole("waiter", "chef", "cashier", "manager", "admin")
}
