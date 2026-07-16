package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Markikie/cinema-booking/internal/ws"
	"github.com/redis/go-redis/v9"
)

const seatEventsChannel = "seat_events"

type PubSubService struct {
	redisClient *redis.Client
	hub         *ws.Hub
}

func NewPubSubService(redisClient *redis.Client, hub *ws.Hub) *PubSubService {
	return &PubSubService{
		redisClient: redisClient,
		hub:         hub,
	}
}

func (p *PubSubService) Publish(ctx context.Context, event ws.SeatEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.redisClient.Publish(ctx, seatEventsChannel, payload).Err()
}

func (p *PubSubService) StartSubscriber(ctx context.Context) {
	pubsub := p.redisClient.Subscribe(ctx, seatEventsChannel)

	go func() {
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			var event ws.SeatEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Println("failed to unmarshal seat event from pubsub:", err)
				continue
			}

			p.hub.Broadcast(event)
		}
	}()

	log.Println("subscribed to redis pubsub channel:", seatEventsChannel)
}
