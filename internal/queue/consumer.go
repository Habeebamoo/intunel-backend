package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Habeebamoo/tunnl-backend/internal/models"
	"github.com/Habeebamoo/tunnl-backend/internal/providers"
	"github.com/redis/go-redis/v9"
)

const (
	ConsumerGroup = "notification-workers"
	ConsumerName  = "worker-1"
)

type Consumer struct {
	client *redis.Client
	router *providers.Router
}

func NewConsumer(client *redis.Client, router *providers.Router) *Consumer {
	return &Consumer{client: client, router: router}
}

func (c *Consumer) Start(ctx context.Context) {
	// Create consumer group if it doesn't exist
	err := c.client.XGroupCreateMkStream(ctx, NotificationStream, ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("failed to create consumer group: %v", err)
	}
	log.Println("Worker: listening on stream...")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			entries, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    ConsumerGroup,
				Consumer: ConsumerName,
				Streams:  []string{NotificationStream, ">"},
				Count:    10,
				Block:    5 * time.Second,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					continue // no messages, keep polling
				}
				log.Printf("stream read error: %v", err)
				continue
			}

			for _, stream := range entries {
				for _, msg := range stream.Messages {
					c.process(ctx, msg)
				}
			}
		}
	}
}

func (c *Consumer) process(ctx context.Context, msg redis.XMessage) {
	raw, ok := msg.Values["data"].(string)
	if !ok {
		log.Printf("invalid message format: %v", msg.ID)
		c.ack(ctx, msg.ID)
		return
	}

	var n models.Notification
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		log.Printf("failed to unmarshal message %s: %v", msg.ID, err)
		c.ack(ctx, msg.ID) // bad message, discard
		return
	}

	log.Printf("processing [%s] → %s", n.Channel, n.To)

	if err := c.router.Route(ctx, n); err != nil {
		log.Printf("failed to send [%s] to %s: %v", n.Channel, n.To, err)
		return // don't ACK — Redis will redeliver
	}

	c.ack(ctx, msg.ID)
}

func (c *Consumer) ack(ctx context.Context, id string) {
	if err := c.client.XAck(ctx, NotificationStream, ConsumerGroup, id).Err(); err != nil {
		log.Printf("failed to ACK message %s: %v", id, err)
	}
}