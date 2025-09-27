package handlers

import (
	"net/http"
	"time"

	"restaurant-system/internal/auth"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthBasicHandler struct{ svc *services.AuthBasicService }

func NewAuthBasicHandler(svc *services.AuthBasicService) *AuthBasicHandler {
	return &AuthBasicHandler{svc: svc}
}

func (h *AuthBasicHandler) Signup(c *gin.Context) {
	var b struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role,omitempty"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Signup(c.Request.Context(), b.Phone, b.Password, b.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *AuthBasicHandler) Signin(c *gin.Context) {
	var b struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, role, err := h.svc.Signin(c.Request.Context(), b.Phone, b.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	// generate JWT token including role
	token, err := auth.GenerateToken(id, role, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "account_id": id, "role": role})
}
