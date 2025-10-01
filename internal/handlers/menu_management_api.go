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

// --- Categories ---
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

func (h *MenuManagementAPI) GetCategory(c *gin.Context) {
	var cat models.MenuCategory
	if err := h.gorm.First(&cat, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

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

func (h *MenuManagementAPI) DeleteCategory(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuCategory{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Items ---
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

func (h *MenuManagementAPI) GetItem(c *gin.Context) {
	var it models.MenuItemGorm
	if err := h.gorm.First(&it, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, it)
}

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

func (h *MenuManagementAPI) DeleteItem(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuItemGorm{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Variants ---
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

func (h *MenuManagementAPI) DeleteVariant(c *gin.Context) {
	if err := h.gorm.Delete(&models.MenuVariant{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Addons ---
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
