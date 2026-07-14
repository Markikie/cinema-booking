package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(uri, dbName string) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping mongodb: %v", err)
	}

	log.Println("connected to mongodb:", uri)
	return client.Database(dbName)
}

func EnsureIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	seatsCol := db.Collection("seats")
	_, err := seatsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{
			"showtime_id": 1,
			"row":         1,
			"number":      1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("warning: failed to create seats index: %v", err)
	}

	bookingsCol := db.Collection("bookings")
	_, err = bookingsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"user_id": 1},
	})
	if err != nil {
		log.Printf("warning: failed to create bookings index: %v", err)
	}

	log.Println("mongodb indexes ensured")
}
