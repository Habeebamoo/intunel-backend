package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/Habeebamoo/intunel-backend/internal/providers"
	"github.com/Habeebamoo/intunel-backend/internal/queue"
	"github.com/redis/go-redis/v9"
)

type Reaper struct {
	client *redis.Client
	router *providers.Router
}

func NewReaper(client *redis.Client, router *providers.Router) *Reaper {
	return &Reaper{client, router}
}

func (r *Reaper) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Reaper: scanning for stuck messages...")
			r.ClaimStuckMessages(ctx)

		case <-ctx.Done():
			return
		}
	}
}

func (r *Reaper) ClaimStuckMessages(ctx context.Context) {
	messages, err := r.FindPendingMessages(ctx)
	if err != nil {
		fmt.Printf("Reaper: error finding pending messages: %v\n", err)
		return
	}

	for _, msg := range messages {
		if msg.Idle < 1 * time.Minute {
			continue
		}

		fmt.Printf("Reaper: claiming message %s (idle for %v)\n", msg.ID, msg.Idle)

		claimedMessages, err := r.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   queue.NotificationStream,
			Group:    queue.ConsumerGroup,
			Consumer: queue.ConsumerName,
			MinIdle:  1 * time.Minute,
			Messages: []string{msg.ID},
		}).Result()

		if err != nil {
			fmt.Printf("Reaper: error claiming message %s: %v\n", msg.ID, err)
		}

		for _, claimed := range claimedMessages {
			go func(m redis.XMessage) {

				r.Process(ctx, m)
			}(claimed)
		}

	}
}

func (r *Reaper) Process(ctx context.Context, msg redis.XMessage) {
	log.Printf("reaper re-processing message %s\n", msg.ID)

	raw, ok := msg.Values["data"].(string)
	if !ok {
		log.Printf("reaper: invalid message format: %v\n", msg.ID)
		r.ack(ctx, msg.ID)
		return
	}

	var n models.Notification
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		log.Printf("reaper: failed to unmarshal message %s: %v\n", msg.ID, err)
		r.ack(ctx, msg.ID)
		return
	}

	log.Printf("reaper: routing [%s] → %s\n", n.Channel, n.To)

	if err := r.router.Route(ctx, n); err != nil {
		log.Printf("reaper: failed to send [%s] to %s: %v\n", n.Channel, n.To, err)
		return
	}
	
	log.Printf("reaper: successfully sent [%s] to %s\n", n.Channel, n.To)
	r.ack(ctx, msg.ID)
}

func (r *Reaper) FindPendingMessages(ctx context.Context) ([]redis.XPendingExt, error) {
	messages, err := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: queue.NotificationStream,
		Group:  queue.ConsumerGroup,
		Start:  "-",
		End:    "+",
		Count:  10,
	}).Result()

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *Reaper) ack(ctx context.Context, id string) {
	if err := r.client.XAck(ctx, queue.NotificationStream, queue.ConsumerGroup, id).Err(); err != nil {
		log.Printf("reaper: failed to ACK message %s: %v\n", id, err)
	}
}