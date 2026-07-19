package repository

import (
	"context"

	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		collection: db.Collection("bookings"),
	}
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) (string, error) {
	booking.ID = primitive.NewObjectID().Hex()
	_, err := r.collection.InsertOne(ctx, booking)
	if err != nil {
		return "", err
	}
	return booking.ID, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, bookingID string, status models.BookingStatus) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": bookingID}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (r *BookingRepository) FindByID(ctx context.Context, bookingID string) (*models.Booking, error) {
	var booking models.Booking
	err := r.collection.FindOne(ctx, bson.M{"_id": bookingID}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) FindPendingBySeat(ctx context.Context, showtimeID, seatID string) ([]models.Booking, error) {
	filter := bson.M{
		"showtime_id": showtimeID,
		"seat_ids":    seatID,
		"status":      models.BookingPending,
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *BookingRepository) FindAll(ctx context.Context, filter bson.M, limit, skip int64) ([]models.Booking, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}
