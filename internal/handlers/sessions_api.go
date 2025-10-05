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

// StartSession godoc
// @Summary Start a table session
// @Description Start a new dining session for a table
// @Tags sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Session true "Session request"
// @Success 201 {object} models.Session
// @@Failure 400 {object} models.ErrorRespons
// @Router /sessions [post]
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

// GetSession godoc
// @Summary Get session by ID
// @Description Retrieve session details
// @Tags sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} models.Session
// @Failure 404 {object} models.ErrorResponse
// @Router /sessions/{id} [get]
func (h *SessionsAPI) GetSession(c *gin.Context) {
	id := c.Param("id")
	s, err := h.svc.GetSession(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// CloseSession godoc
// @Summary Close a session
// @Description Close an active dining session
// @Tags sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 "Session closed"
// @@Failure 400 {object} models.ErrorRespons
// @Router /sessions/{id}/close [put]
func (h *SessionsAPI) CloseSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CloseSession(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
