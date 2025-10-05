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

// Signup godoc
// @Summary Sign up a new user
// @Description Create a new user account with phone and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{phone=string,password=string,role=string} true "Signup request"
// @Success 201 "Account created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Router /auth/signup [post]
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

// Signin godoc
// @Summary Sign in to an existing account
// @Description Authenticate with phone and password to receive a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{phone=string,password=string} true "Signin request"
// @Success 200 {object} models.TokenResponse "Returns token, account_id, and role"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials"
// @Router /auth/signin [post]
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
