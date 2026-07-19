package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Markikie/cinema-booking/internal/config"
	"github.com/Markikie/cinema-booking/internal/database"
	"github.com/Markikie/cinema-booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type seedShowtime struct {
	ID          string
	MovieName   string
	Hall        string
	StartOffset time.Duration
	Rows        int
	SeatsPerRow int
}

func main() {
	cfg := config.Load()
	db := database.NewMongoClient(cfg.MongoURI, cfg.MongoDBName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	showtimes := []seedShowtime{
		{ID: "seed-avatar-3", MovieName: "Avatar 3", Hall: "Hall A", StartOffset: 24 * time.Hour, Rows: 6, SeatsPerRow: 10},
		{ID: "seed-dune-part-three", MovieName: "Dune: Part Three", Hall: "Hall B", StartOffset: 26 * time.Hour, Rows: 6, SeatsPerRow: 10},
		{ID: "seed-spider-man-4", MovieName: "Spider-Man 4", Hall: "Hall C", StartOffset: 28 * time.Hour, Rows: 5, SeatsPerRow: 8},
		{ID: "seed-the-batman-part-ii", MovieName: "The Batman Part II", Hall: "Hall A", StartOffset: 48 * time.Hour, Rows: 6, SeatsPerRow: 10},
		{ID: "seed-inside-out-3", MovieName: "Inside Out 3", Hall: "Hall B", StartOffset: 50 * time.Hour, Rows: 5, SeatsPerRow: 8},
		{ID: "seed-jurassic-world-rebirth", MovieName: "Jurassic World Rebirth", Hall: "Hall C", StartOffset: 52 * time.Hour, Rows: 6, SeatsPerRow: 10},
		{ID: "seed-mission-impossible-finale", MovieName: "Mission: Impossible Finale", Hall: "Hall A", StartOffset: 72 * time.Hour, Rows: 6, SeatsPerRow: 10},
		{ID: "seed-fantastic-four", MovieName: "The Fantastic Four", Hall: "Hall B", StartOffset: 74 * time.Hour, Rows: 5, SeatsPerRow: 8},
		{ID: "seed-wicked-part-two", MovieName: "Wicked: Part Two", Hall: "Hall C", StartOffset: 76 * time.Hour, Rows: 5, SeatsPerRow: 8},
		{ID: "seed-tron-ares", MovieName: "Tron: Ares", Hall: "Hall D", StartOffset: 78 * time.Hour, Rows: 6, SeatsPerRow: 10},
	}

	showtimeCollection := db.Collection("showtimes")
	seatCollection := db.Collection("seats")
	now := time.Now().Truncate(time.Minute)

	totalSeats := 0
	for _, item := range showtimes {
		showtime := models.Showtime{
			ID:          item.ID,
			MovieName:   item.MovieName,
			Hall:        item.Hall,
			StartTime:   now.Add(item.StartOffset),
			Rows:        item.Rows,
			SeatsPerRow: item.SeatsPerRow,
		}

		_, err := showtimeCollection.UpdateOne(
			ctx,
			bson.M{"_id": showtime.ID},
			bson.M{"$set": showtime},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			log.Fatalf("failed to seed showtime %s: %v", showtime.MovieName, err)
		}

		for row := 0; row < item.Rows; row++ {
			rowLabel := string(rune('A' + row))
			for number := 1; number <= item.SeatsPerRow; number++ {
				seat := models.Seat{
					ID:         fmt.Sprintf("%s-%s-%02d", item.ID, rowLabel, number),
					ShowtimeID: item.ID,
					Row:        rowLabel,
					Number:     number,
					Status:     models.SeatAvailable,
				}

				_, err := seatCollection.UpdateOne(
					ctx,
					bson.M{"_id": seat.ID},
					bson.M{"$setOnInsert": seat},
					options.Update().SetUpsert(true),
				)
				if err != nil {
					log.Fatalf("failed to seed seat %s: %v", seat.ID, err)
				}
				totalSeats++
			}
		}
	}

	log.Printf("seeded %d showtimes and ensured %d seats", len(showtimes), totalSeats)
}
