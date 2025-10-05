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

// CreateReservation godoc
// @Summary Create a reservation
// @Description Create a new table reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Param request body models.Reservation true "Reservation request"
// @Success 201 {object} models.Reservation
// @@Failure 400 {object} models.ErrorRespons
// @Router /reservations [post]
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

// ListReservations godoc
// @Summary List reservations
// @Description Get all upcoming reservations
// @Tags reservations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /reservations [get]
func (h *ReservationsAPI) ListReservations(c *gin.Context) {
	res, err := h.svc.ListUpcoming(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reservations": res})
}

// UpdateReservation godoc
// @Summary Update a reservation
// @Description Update reservation details
// @Tags reservations
// @Accept json
// @Produce json
// @Param id path string true "Reservation ID"
// @Param request body models.Reservation true "Reservation update"
// @Success 200 {object} models.Reservation
// @@Failure 400 {object} models.ErrorRespons
// @Router /reservations/{id} [put]
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

// CancelReservation godoc
// @Summary Cancel a reservation
// @Description Cancel an existing reservation
// @Tags reservations
// @Produce json
// @Param id path string true "Reservation ID"
// @Success 200 "Reservation cancelled"
// @Failure 500 {object} models.ErrorResponse
// @Router /reservations/{id} [delete]
func (h *ReservationsAPI) CancelReservation(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CancelReservation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
