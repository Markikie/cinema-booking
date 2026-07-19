package repository

import (
	"context"
	"time"

	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SeatRepository struct {
	collection *mongo.Collection
}

func NewSeatRepository(db *mongo.Database) *SeatRepository {
	return &SeatRepository{
		collection: db.Collection("seats"),
	}
}

func (r *SeatRepository) FindByShowtime(ctx context.Context, showtimeID string) ([]models.Seat, error) {
	opts := options.Find().SetSort(bson.D{{Key: "row", Value: 1}, {Key: "number", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"showtime_id": showtimeID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var seats []models.Seat
	if err := cursor.All(ctx, &seats); err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *SeatRepository) FindByID(ctx context.Context, seatID string) (*models.Seat, error) {
	objID, err := primitive.ObjectIDFromHex(seatID)
	if err != nil {
		return nil, err
	}

	var seat models.Seat
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&seat)
	if err != nil {
		return nil, err
	}
	return &seat, nil
}

func (r *SeatRepository) UpdateStatus(ctx context.Context, seatID string, status models.SeatStatus, lockedBy string) error {
	objID, err := primitive.ObjectIDFromHex(seatID)
	if err != nil {
		return err
	}

	update := bson.M{
		"status": status,
	}

	if status == models.SeatLocked {
		now := time.Now()
		update["locked_by"] = lockedBy
		update["locked_at"] = now
	} else {
		update["locked_by"] = ""
		update["locked_at"] = nil
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	return err
}

func (r *SeatRepository) MarkBookedIfLockedBy(ctx context.Context, seatID, lockedBy string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(seatID)
	if err != nil {
		return false, err
	}

	filter := bson.M{
		"_id":       objID,
		"status":    models.SeatLocked,
		"locked_by": lockedBy,
	}
	update := bson.M{
		"$set": bson.M{
			"status":    models.SeatBooked,
			"locked_by": "",
			"locked_at": nil,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}

func (r *SeatRepository) InsertMany(ctx context.Context, seats []interface{}) error {
	_, err := r.collection.InsertMany(ctx, seats)
	return err
}
