package models

import "time"

type BookingStatus string

const (
	BookingPending BookingStatus = "PENDING"
	BookingSuccess BookingStatus = "SUCCESS"
	BookingTimeout BookingStatus = "TIMEOUT"
	BookingFailed  BookingStatus = "FAILED"
)

type Booking struct {
	ID         string        `bson:"_id,omitempty" json:"id"`
	UserID     string        `bson:"user_id" json:"user_id"`
	ShowtimeID string        `bson:"showtime_id" json:"showtime_id"`
	SeatIDs    []string      `bson:"seat_ids" json:"seat_ids"`
	Status     BookingStatus `bson:"status" json:"status"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`

	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
}

type AuditLog struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	EventType string    `bson:"event_type" json:"event_type"` // "BOOKING_SUCCESS", "BOOKING_TIMEOUT", "SEAT_RELEASED", "SYSTEM_ERROR"
	UserID    string    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	BookingID string    `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
	Detail    string    `bson:"detail" json:"detail"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
