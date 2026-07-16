package handlers

import (
	"net/http"
	"strconv"

	"github.com/Markikie/cinema-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type AdminHandler struct {
	bookingRepo *repository.BookingRepository
	auditRepo   *repository.AuditLogRepository
}

func NewAdminHandler(bookingRepo *repository.BookingRepository, auditRepo *repository.AuditLogRepository) *AdminHandler {
	return &AdminHandler{
		bookingRepo: bookingRepo,
		auditRepo:   auditRepo,
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
