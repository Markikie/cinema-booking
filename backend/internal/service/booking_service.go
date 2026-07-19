package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Markikie/cinema-booking/internal/models"
	"github.com/Markikie/cinema-booking/internal/repository"
	"github.com/Markikie/cinema-booking/internal/ws"
)

var (
	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatNotBookable   = errors.New("seat is not available for booking")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrBookingExpired    = errors.New("booking has expired")
	ErrSeatNotLockedByMe = errors.New("seat is not locked by this user")
	ErrBookingNotOwned   = errors.New("booking does not belong to this user")
	ErrSeatConflict      = errors.New("seat could not be confirmed, it may already be booked or its lock expired")
)

type BookingService struct {
	seatRepo    *repository.SeatRepository
	bookingRepo *repository.BookingRepository
	auditRepo   *repository.AuditLogRepository
	lockService *LockService
	pubsub      *PubSubService
	lockTTL     time.Duration
}

func NewBookingService(
	seatRepo *repository.SeatRepository,
	bookingRepo *repository.BookingRepository,
	auditRepo *repository.AuditLogRepository,
	lockService *LockService,
	pubsub *PubSubService,
	lockTTL time.Duration,
) *BookingService {
	return &BookingService{
		seatRepo:    seatRepo,
		bookingRepo: bookingRepo,
		auditRepo:   auditRepo,
		lockService: lockService,
		pubsub:      pubsub,
		lockTTL:     lockTTL,
	}
}

func (s *BookingService) SelectSeat(ctx context.Context, userID, showtimeID, seatID string) error {
	seat, err := s.seatRepo.FindByID(ctx, seatID)
	if err != nil {
		return ErrSeatNotFound
	}

	if seat.Status == models.SeatBooked {
		return ErrSeatNotBookable
	}

	if err := s.lockService.AcquireLock(ctx, showtimeID, seatID, userID); err != nil {
		if errors.Is(err, ErrLockNotAcquired) {
			return ErrSeatNotBookable
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if err := s.seatRepo.UpdateStatus(ctx, seatID, models.SeatLocked, userID); err != nil {
		_ = s.lockService.ReleaseLock(ctx, showtimeID, seatID, userID) // best-effort rollback
		s.logEvent(ctx, "SYSTEM_ERROR", userID, "", fmt.Sprintf("failed to update seat status after lock: %v", err))
		return fmt.Errorf("failed to update seat status: %w", err)
	}

	s.logEvent(ctx, "SEAT_LOCKED", userID, "", fmt.Sprintf("seat %s locked for showtime %s", seatID, showtimeID))

	_ = s.pubsub.Publish(ctx, ws.SeatEvent{
		Type:       "SEAT_LOCKED",
		ShowtimeID: showtimeID,
		SeatID:     seatID,
		Status:     string(models.SeatLocked),
	})

	return nil
}

// CreateBooking เดิมไม่เคยเช็คเลยว่าที่นั่งที่ขอจองถูกล็อกโดย user คนนี้จริงไหม
// (client ข้ามขั้นตอน select-seat แล้วยิงมาตรงนี้พร้อม seat_id ใดก็ได้ก็ผ่าน) —
// เพิ่มการเช็คว่าทุกที่นั่งต้องอยู่ในสถานะ LOCKED และ locked_by ต้องเป็น userID
// นี้เท่านั้น ก่อนจะสร้าง booking record เพื่อให้ "การล็อก" มีความหมายจริง
// ไม่ใช่แค่ UI hint
func (s *BookingService) CreateBooking(ctx context.Context, userID, showtimeID string, seatIDs []string) (string, error) {
	for _, seatID := range seatIDs {
		seat, err := s.seatRepo.FindByID(ctx, seatID)
		if err != nil {
			return "", ErrSeatNotFound
		}
		if seat.Status != models.SeatLocked || seat.LockedBy != userID {
			return "", ErrSeatNotLockedByMe
		}
	}

	now := time.Now()
	booking := &models.Booking{
		UserID:     userID,
		ShowtimeID: showtimeID,
		SeatIDs:    seatIDs,
		Status:     models.BookingPending,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.lockTTL),
	}

	bookingID, err := s.bookingRepo.Create(ctx, booking)
	if err != nil {
		return "", fmt.Errorf("failed to create booking: %w", err)
	}

	return bookingID, nil
}

func (s *BookingService) ConfirmPayment(ctx context.Context, bookingID, userID string) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if booking.UserID != userID {
		return ErrBookingNotOwned
	}

	if time.Now().After(booking.ExpiresAt) {
		return ErrBookingExpired
	}

	for _, seatID := range booking.SeatIDs {
		ok, err := s.seatRepo.MarkBookedIfLockedBy(ctx, seatID, userID)
		if err != nil {
			s.logEvent(ctx, "SYSTEM_ERROR", userID, bookingID, fmt.Sprintf("failed to mark seat booked: %v", err))
			return fmt.Errorf("failed to update seat status: %w", err)
		}
		if !ok {
			s.logEvent(ctx, "SYSTEM_ERROR", userID, bookingID, fmt.Sprintf("seat %s conflict during confirm-payment", seatID))
			return ErrSeatConflict
		}

		_ = s.lockService.ReleaseLock(ctx, booking.ShowtimeID, seatID, userID)

		_ = s.pubsub.Publish(ctx, ws.SeatEvent{
			Type:       "SEAT_BOOKED",
			ShowtimeID: booking.ShowtimeID,
			SeatID:     seatID,
			Status:     string(models.SeatBooked),
		})
	}

	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, models.BookingSuccess); err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	s.logEvent(ctx, "BOOKING_SUCCESS", userID, bookingID, fmt.Sprintf("booking %s confirmed for showtime %s", bookingID, booking.ShowtimeID))
	return nil
}

func (s *BookingService) ReleaseExpiredSeat(ctx context.Context, showtimeID, seatID string) {
	seat, err := s.seatRepo.FindByID(ctx, seatID)
	if err != nil {
		log.Println("failed to find seat during expiry release:", err)
		return
	}

	if seat.Status != models.SeatLocked {
		return
	}

	if err := s.seatRepo.UpdateStatus(ctx, seatID, models.SeatAvailable, ""); err != nil {
		log.Println("failed to release expired seat:", err)
		return
	}

	_ = s.pubsub.Publish(ctx, ws.SeatEvent{
		Type:       "SEAT_RELEASED",
		ShowtimeID: showtimeID,
		SeatID:     seatID,
		Status:     string(models.SeatAvailable),
	})

	s.logEvent(ctx, "BOOKING_TIMEOUT", seat.LockedBy, "", fmt.Sprintf("seat %s released after lock expired", seatID))
	s.logEvent(ctx, "SEAT_RELEASED", seat.LockedBy, "", fmt.Sprintf("seat %s is now available", seatID))

	pendingBookings, err := s.bookingRepo.FindPendingBySeat(ctx, showtimeID, seatID)
	if err != nil {
		log.Println("failed to find pending bookings during expiry release:", err)
		return
	}
	for _, b := range pendingBookings {
		if err := s.bookingRepo.UpdateStatus(ctx, b.ID, models.BookingTimeout); err != nil {
			log.Println("failed to mark booking as timeout:", err)
		}
	}
}

func (s *BookingService) logEvent(ctx context.Context, eventType, userID, bookingID, detail string) {
	err := s.auditRepo.Create(ctx, &models.AuditLog{
		EventType: eventType,
		UserID:    userID,
		BookingID: bookingID,
		Detail:    detail,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("warning: failed to write audit log:", err)
	}
}
