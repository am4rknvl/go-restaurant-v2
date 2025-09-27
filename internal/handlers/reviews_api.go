package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type ReviewsAPI struct{ svc *services.MenuSQLService }

func NewReviewsAPI(svc *services.MenuSQLService) *ReviewsAPI { return &ReviewsAPI{svc: svc} }

// POST /api/v1/reviews
func (h *ReviewsAPI) CreateReview(c *gin.Context) {
	var body struct {
		MenuItemID string `json:"menu_item_id" binding:"required"`
		Rating     int    `json:"rating" binding:"required"`
		Comment    string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accountID := c.GetString("account_id")
	r := &models.Review{ID: uuid.New().String(), AccountID: accountID, MenuItemID: body.MenuItemID, Rating: body.Rating, Comment: body.Comment}
	if err := h.svc.CreateReview(c.Request.Context(), r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}
