package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type FavoritesAPI struct{ svc *services.MenuSQLService }

func NewFavoritesAPI(svc *services.MenuSQLService) *FavoritesAPI { return &FavoritesAPI{svc: svc} }

// AddFavorite godoc
// @Summary Add favorite menu item
// @Description Add a menu item to user's favorites
// @Tags favorites
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{menu_item_id=string} true "Favorite request"
// @Success 201 {object} models.Favorite
// @@Failure 400 {object} models.ErrorRespons
// @Router /favorites [post]
func (h *FavoritesAPI) AddFavorite(c *gin.Context) {
	var body struct {
		MenuItemID string `json:"menu_item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accountID := c.GetString("account_id")
	fav := &models.Favorite{ID: uuid.New().String(), AccountID: accountID, MenuItemID: body.MenuItemID}
	if err := h.svc.AddFavorite(c.Request.Context(), fav); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, fav)
}

// RemoveFavorite godoc
// @Summary Remove favorite menu item
// @Description Remove a menu item from user's favorites
// @Tags favorites
// @Produce json
// @Security BearerAuth
// @Param menu_item_id path string true "Menu Item ID"
// @Success 204 "Removed"
// @Failure 500 {object} models.ErrorResponse
// @Router /favorites/{menu_item_id} [delete]
func (h *FavoritesAPI) RemoveFavorite(c *gin.Context) {
	menuItemID := c.Param("menu_item_id")
	accountID := c.GetString("account_id")
	if err := h.svc.RemoveFavorite(c.Request.Context(), accountID, menuItemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
