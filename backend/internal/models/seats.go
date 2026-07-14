package models

import "time"

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

type Seat struct {
	ID         string     `bson:"_id,omitempty" json:"id"`
	ShowtimeID string     `bson:"showtime_id" json:"showtime_id"`
	Row        string     `bson:"row" json:"row"`
	Number     int        `bson:"number" json:"number"`
	Status     SeatStatus `bson:"status" json:"status"`
	LockedBy   string     `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
	LockedAt   *time.Time `bson:"locked_at,omitempty" json:"locked_at,omitempty"`
}

// Showtime คือรอบฉายหนังหนึ่งรอบ
type Showtime struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	MovieName   string    `bson:"movie_name" json:"movie_name"`
	Hall        string    `bson:"hall" json:"hall"`
	StartTime   time.Time `bson:"start_time" json:"start_time"`
	Rows        int       `bson:"rows" json:"rows"`
	SeatsPerRow int       `bson:"seats_per_row" json:"seats_per_row"`
}
