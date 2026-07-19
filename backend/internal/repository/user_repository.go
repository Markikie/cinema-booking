package repository

import (
	"context"
	"strings"
	"time"

	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection  *mongo.Collection
	adminEmails map[string]bool
}

func NewUserRepository(db *mongo.Database, adminEmails []string) *UserRepository {
	lookup := make(map[string]bool, len(adminEmails))
	for _, e := range adminEmails {
		lookup[strings.ToLower(e)] = true
	}
	return &UserRepository{
		collection:  db.Collection("users"),
		adminEmails: lookup,
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
	desiredRole := models.RoleUser
	if r.adminEmails[strings.ToLower(email)] {
		desiredRole = models.RoleAdmin
	}

	existing, err := r.FindByGoogleID(ctx, googleID)
	if err == nil {
		if existing.Role != desiredRole {
			_, _ = r.collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": bson.M{"role": desiredRole}})
			existing.Role = desiredRole
		}
		return existing, nil
	}

	newUser := &models.User{
		ID:        primitive.NewObjectID().Hex(),
		GoogleID:  googleID,
		Email:     email,
		Name:      name,
		Role:      desiredRole,
		CreatedAt: time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}
