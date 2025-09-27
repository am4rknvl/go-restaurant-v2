package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SessionsAPI struct {
	svc *services.SessionService
}

func NewSessionsAPI(svc *services.SessionService) *SessionsAPI { return &SessionsAPI{svc: svc} }

// POST /api/v1/sessions
func (h *SessionsAPI) StartSession(c *gin.Context) {
	var body models.Session
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	s, err := h.svc.StartSession(c.Request.Context(), &body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// GET /api/v1/sessions/:id
func (h *SessionsAPI) GetSession(c *gin.Context) {
	id := c.Param("id")
	s, err := h.svc.GetSession(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// PUT /api/v1/sessions/:id/close
func (h *SessionsAPI) CloseSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CloseSession(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
