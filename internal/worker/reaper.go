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

const (
	retryHashPrefix = "notifications:retry:"
	dlqStream       = "notifications:stream:dead"
	maxRetries      = 3
)

type Reaper struct {
	client *redis.Client
	router *providers.Router
}

func NewReaper(client *redis.Client, router *providers.Router) *Reaper {
	return &Reaper{client, router}
}

func (r *Reaper) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
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
			count := r.getRetryCount(ctx, msg.ID)

			switch count {
			case 0:
					r.claimAndProcess(ctx, msg.ID, 1*time.Minute)
			case 1:
					if msg.Idle >= 1*time.Minute {
						r.claimAndProcess(ctx, msg.ID, 1*time.Minute)
					}
			case 2:
					if msg.Idle >= 5*time.Minute {
						r.claimAndProcess(ctx, msg.ID, 5*time.Minute)
					}
			default:
					r.claimAndProcess(ctx, msg.ID, 1*time.Minute)
			}
	}
}

func (r *Reaper) claimAndProcess(ctx context.Context, msgID string, minIdle time.Duration) {
	claimedMessages, err := r.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   queue.NotificationStream,
			Group:    queue.ConsumerGroup,
			Consumer: queue.ConsumerName,
			MinIdle:  minIdle,
			Messages: []string{msgID},
	}).Result()

	if err != nil || len(claimedMessages) == 0 {
    fmt.Printf("Reaper: message %s not ready yet or already processed\n", msgID)
    return
	}

	go func(m redis.XMessage) {
		r.Process(ctx, m)
	}(claimedMessages[0])
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

	if err := r.router.Route(ctx, n); err != nil {
		log.Printf("reaper: failed to send [%s] to %s: %v\n", n.Channel, n.To, err)
		r.handleFailure(ctx, msg, err.Error())
		return
	}

	log.Printf("reaper: successfully sent [%s] to %s\n", n.Channel, n.To)
	r.client.Del(ctx, retryHashPrefix+msg.ID)
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

func (r *Reaper) getRetryCount(ctx context.Context, msgID string) int {
	val, err := r.client.HGet(ctx, retryHashPrefix+msgID, "count").Int()
	if err != nil {
			return 0
	}
	return val
}

func (r *Reaper) handleFailure(ctx context.Context, msg redis.XMessage, errReason string) {
	key := retryHashPrefix + msg.ID
	count := r.getRetryCount(ctx, msg.ID)

	if count == 0 {
		r.client.HSet(ctx, key, "first_failed_at", time.Now().Unix())
	}

	newCount := count + 1
	r.client.HSet(ctx, key, "count", newCount)
	r.client.Expire(ctx, key, 24*time.Hour)

	if newCount >= maxRetries {
		log.Printf("reaper: message %s exceeded max retries, moving to DLQ\n", msg.ID)
		r.sendToDLQ(ctx, msg, errReason)
		return
	}

	log.Printf("reaper: message %s failed, retry count now %d\n", msg.ID, newCount)
}

func (r *Reaper) sendToDLQ(ctx context.Context, msg redis.XMessage, errReason string) {
	raw, _ := r.client.XRange(ctx, queue.NotificationStream, msg.ID, msg.ID).Result()

	originalData := ""
	if len(raw) > 0 {
			if data, ok := raw[0].Values["data"].(string); ok {
					originalData = data
			}
	}

	r.client.XAdd(ctx, &redis.XAddArgs{
			Stream: dlqStream,
			Values: map[string]interface{}{
					"data":      originalData,
					"error":     errReason,
					"failed_at": time.Now().Unix(),
					"msg_id":    msg.ID,
			},
	})

	r.client.Del(ctx, retryHashPrefix+msg.ID)
	r.ack(ctx, msg.ID)
	log.Printf("reaper: message %s moved to DLQ\n", msg.ID)
}

func (r *Reaper) ack(ctx context.Context, id string) {
	if err := r.client.XAck(ctx, queue.NotificationStream, queue.ConsumerGroup, id).Err(); err != nil {
		log.Printf("reaper: failed to ACK message %s: %v\n", id, err)
	}
}