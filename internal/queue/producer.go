package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Habeebamoo/tunnl-backend/internal/models"
	"github.com/redis/go-redis/v9"
)

const NotificationStream = "notifications:stream"

type Producer struct {
	client *redis.Client
}

func NewProducer(client *redis.Client) *Producer {
	return &Producer{client: client}
}

func (p *Producer) Publish(ctx context.Context, n models.Notification) error {
	payload, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: NotificationStream,
		Values: map[string]interface{}{
			"data": payload,
		},
	}).Err()
}