package handlers

import (
	"errors"
	"net/http"

	"github.com/Markikie/cinema-booking/internal/middleware"
	"github.com/Markikie/cinema-booking/internal/repository"
	"github.com/Markikie/cinema-booking/internal/service"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService *service.BookingService
	seatRepo       *repository.SeatRepository
}

func NewBookingHandler(bookingService *service.BookingService, seatRepo *repository.SeatRepository) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		seatRepo:       seatRepo,
	}
}

func (h *BookingHandler) GetSeatMap(c *gin.Context) {
	showtimeID := c.Param("showtime_id")

	seats, err := h.seatRepo.FindByShowtime(c.Request.Context(), showtimeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch seat map"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"seats": seats})
}

type selectSeatRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatID     string `json:"seat_id" binding:"required"`
}

func (h *BookingHandler) SelectSeat(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	var req selectSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id and seat_id are required"})
		return
	}

	err := h.bookingService.SelectSeat(c.Request.Context(), userID, req.ShowtimeID, req.SeatID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSeatNotBookable):
			c.JSON(http.StatusConflict, gin.H{"error": "seat is already locked or booked"})
		case errors.Is(err, service.ErrSeatNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "seat not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to select seat"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "seat locked successfully", "lock_duration_seconds": 300})
}

type releaseSeatRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatID     string `json:"seat_id" binding:"required"`
}

func (h *BookingHandler) ReleaseSeat(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	var req releaseSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id and seat_id are required"})
		return
	}

	err := h.bookingService.ReleaseSeat(c.Request.Context(), userID, req.ShowtimeID, req.SeatID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSeatNotLockedByMe):
			c.JSON(http.StatusConflict, gin.H{"error": "seat is not locked by you"})
		case errors.Is(err, service.ErrSeatNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "seat not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release seat"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "seat released successfully"})
}

type createBookingRequest struct {
	ShowtimeID string   `json:"showtime_id" binding:"required"`
	SeatIDs    []string `json:"seat_ids" binding:"required,min=1"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime_id and at least one seat_id are required"})
		return
	}

	bookingID, err := h.bookingService.CreateBooking(c.Request.Context(), userID, req.ShowtimeID, req.SeatIDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSeatNotLockedByMe):

			c.JSON(http.StatusConflict, gin.H{"error": "one or more seats are not locked by you — select them again"})
		case errors.Is(err, service.ErrSeatNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "seat not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"booking_id": bookingID})
}

type confirmPaymentRequest struct {
	BookingID string `json:"booking_id" binding:"required"`
}

func (h *BookingHandler) ConfirmPayment(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	var req confirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id is required"})
		return
	}

	err := h.bookingService.ConfirmPayment(c.Request.Context(), req.BookingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		case errors.Is(err, service.ErrBookingNotOwned):

			c.JSON(http.StatusForbidden, gin.H{"error": "this booking does not belong to you"})
		case errors.Is(err, service.ErrBookingExpired):
			c.JSON(http.StatusGone, gin.H{"error": "booking has expired, please select seats again"})
		case errors.Is(err, service.ErrSeatConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "one or more seats could not be confirmed, please select seats again"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm payment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking confirmed"})
}
