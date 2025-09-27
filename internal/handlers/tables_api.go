package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TablesAPI struct {
	svc *services.TableService
}

func NewTablesAPI(svc *services.TableService) *TablesAPI { return &TablesAPI{svc: svc} }

// POST /api/v1/tables
func (h *TablesAPI) CreateTable(c *gin.Context) {
	var body models.Table
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ID == "" {
		body.ID = uuid.New().String()
	}
	if err := h.svc.CreateTable(c.Request.Context(), &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, body)
}

// GET /api/v1/tables
func (h *TablesAPI) ListTables(c *gin.Context) {
	res, err := h.svc.ListTables(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tables": res})
}

// GET /api/v1/tables/:id
func (h *TablesAPI) GetTable(c *gin.Context) {
	id := c.Param("id")
	t, err := h.svc.GetTable(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}
