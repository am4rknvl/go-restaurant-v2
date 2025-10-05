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

// CreateTable godoc
// @Summary Create a new table
// @Description Create a new restaurant table
// @Tags tables
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Table true "Table request"
// @Success 201 {object} models.Table
// @@Failure 400 {object} models.ErrorRespons
// @Router /tables [post]
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

// ListTables godoc
// @Summary List all tables
// @Description Get a list of all restaurant tables
// @Tags tables
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /tables [get]
func (h *TablesAPI) ListTables(c *gin.Context) {
	res, err := h.svc.ListTables(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tables": res})
}

// GetTable godoc
// @Summary Get table by ID
// @Description Retrieve table details by ID
// @Tags tables
// @Produce json
// @Security BearerAuth
// @Param id path string true "Table ID"
// @Success 200 {object} models.Table
// @Failure 404 {object} models.ErrorResponse
// @Router /tables/{id} [get]
func (h *TablesAPI) GetTable(c *gin.Context) {
	id := c.Param("id")
	t, err := h.svc.GetTable(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}
