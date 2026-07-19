package repository

import (
	"context"

	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShowtimeRepository struct {
	collection *mongo.Collection
}

func NewShowtimeRepository(db *mongo.Database) *ShowtimeRepository {
	return &ShowtimeRepository{
		collection: db.Collection("showtimes"),
	}
}

func (r *ShowtimeRepository) Create(ctx context.Context, showtime *models.Showtime) (string, error) {
	res, err := r.collection.InsertOne(ctx, showtime)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *ShowtimeRepository) FindByID(ctx context.Context, showtimeID string) (*models.Showtime, error) {
	objID, err := primitive.ObjectIDFromHex(showtimeID)
	if err != nil {
		return nil, err
	}
	var showtime models.Showtime
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&showtime)
	if err != nil {
		return nil, err
	}
	return &showtime, nil
}

func (r *ShowtimeRepository) FindAll(ctx context.Context) ([]models.Showtime, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var showtimes []models.Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, err
	}
	return showtimes, nil
}
