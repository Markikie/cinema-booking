package service

import (
	"context"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

type ExpiryListener struct {
	redisClient    *redis.Client
	bookingService *BookingService
}

func NewExpiryListener(redisClient *redis.Client, bookingService *BookingService) *ExpiryListener {
	return &ExpiryListener{
		redisClient:    redisClient,
		bookingService: bookingService,
	}
}

func (l *ExpiryListener) Start(ctx context.Context) {
	pubsub := l.redisClient.Subscribe(ctx, "__keyevent@0__:expired")

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()

		for msg := range ch {
			key := msg.Payload
			if !strings.HasPrefix(key, "seat_lock:") {
				continue
			}

			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				log.Println("unexpected seat_lock key format:", key)
				continue
			}
			showtimeID, seatID := parts[1], parts[2]

			l.bookingService.ReleaseExpiredSeat(context.Background(), showtimeID, seatID)
		}
	}()

	log.Println("subscribed to redis keyspace expiry notifications")
}

func EnableKeyspaceNotifications(ctx context.Context, redisClient *redis.Client) error {
	return redisClient.ConfigSet(ctx, "notify-keyspace-events", "Ex").Err()
}
