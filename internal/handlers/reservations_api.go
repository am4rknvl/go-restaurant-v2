package handlers

import (
	"net/http"

	"restaurant-system/internal/models"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

type ReservationsAPI struct {
	svc *services.ReservationService
}

func NewReservationsAPI(svc *services.ReservationService) *ReservationsAPI {
	return &ReservationsAPI{svc: svc}
}

// POST /api/v1/reservations
func (h *ReservationsAPI) CreateReservation(c *gin.Context) {
	var r models.Reservation
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateReservation(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

// GET /api/v1/reservations
func (h *ReservationsAPI) ListReservations(c *gin.Context) {
	res, err := h.svc.ListUpcoming(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservations": res})
}

// PUT /api/v1/reservations/:id
func (h *ReservationsAPI) UpdateReservation(c *gin.Context) {
	id := c.Param("id")
	var r models.Reservation
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.ID = id
	if err := h.svc.UpdateReservation(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

// DELETE /api/v1/reservations/:id
func (h *ReservationsAPI) CancelReservation(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CancelReservation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
