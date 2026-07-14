package repository

import (
	"context"
	"time"

	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"google_id": googleID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindOrCreate(ctx context.Context, googleID, email, name string) (*models.User, error) {
	existing, err := r.FindByGoogleID(ctx, googleID)
	if err == nil {
		return existing, nil
	}

	newUser := &models.User{
		GoogleID:  googleID,
		Email:     email,
		Name:      name,
		Role:      models.RoleUser,
		CreatedAt: time.Now(),
	}

	res, err := r.collection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	newUser.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return newUser, nil
}
