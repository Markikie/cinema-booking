package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Markikie/cinema-booking/internal/models"
	"github.com/Markikie/cinema-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type AdminHandler struct {
	bookingRepo  *repository.BookingRepository
	auditRepo    *repository.AuditLogRepository
	showtimeRepo *repository.ShowtimeRepository
	seatRepo     *repository.SeatRepository
}

func NewAdminHandler(
	bookingRepo *repository.BookingRepository,
	auditRepo *repository.AuditLogRepository,
	showtimeRepo *repository.ShowtimeRepository,
	seatRepo *repository.SeatRepository,
) *AdminHandler {
	return &AdminHandler{
		bookingRepo:  bookingRepo,
		auditRepo:    auditRepo,
		showtimeRepo: showtimeRepo,
		seatRepo:     seatRepo,
	}
}

func (h *AdminHandler) ListBookings(c *gin.Context) {
	filter := bson.M{}

	if showtimeID := c.Query("showtime_id"); showtimeID != "" {
		filter["showtime_id"] = showtimeID
	}
	if userID := c.Query("user_id"); userID != "" {
		filter["user_id"] = userID
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	limit := parseQueryInt(c.Query("limit"), 50)
	skip := parseQueryInt(c.Query("skip"), 0)

	bookings, err := h.bookingRepo.FindAll(c.Request.Context(), filter, int64(limit), int64(skip))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	limit := parseQueryInt(c.Query("limit"), 100)

	logs, err := h.auditRepo.FindAll(c.Request.Context(), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

type createShowtimeRequest struct {
	MovieName   string    `json:"movie_name" binding:"required"`
	Hall        string    `json:"hall" binding:"required"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	Rows        int       `json:"rows" binding:"required,min=1"`
	SeatsPerRow int       `json:"seats_per_row" binding:"required,min=1"`
}

func (h *AdminHandler) CreateShowtime(c *gin.Context) {
	var req createShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showtime := &models.Showtime{
		MovieName:   req.MovieName,
		Hall:        req.Hall,
		StartTime:   req.StartTime,
		Rows:        req.Rows,
		SeatsPerRow: req.SeatsPerRow,
	}

	showtimeID, err := h.showtimeRepo.Create(c.Request.Context(), showtime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create showtime"})
		return
	}

	seats := make([]interface{}, 0, req.Rows*req.SeatsPerRow)
	for row := 0; row < req.Rows; row++ {
		rowLabel := string(rune('A' + row))
		for number := 1; number <= req.SeatsPerRow; number++ {
			seats = append(seats, models.Seat{
				ShowtimeID: showtimeID,
				Row:        rowLabel,
				Number:     number,
				Status:     models.SeatAvailable,
			})
		}
	}
	if err := h.seatRepo.InsertMany(c.Request.Context(), seats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "showtime created but failed to seed seats"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"showtime_id": showtimeID, "seats_created": len(seats)})
}

func parseQueryInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return n
}
