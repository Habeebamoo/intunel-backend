package services

import (
	"context"
	"fmt"

	"github.com/Habeebamoo/tunnl-backend/internal/models"
	"github.com/Habeebamoo/tunnl-backend/internal/queue"
)

type NotificationService interface {
	SendNotification(ctx context.Context, n models.Notification) error
}

type notificationService struct {
	producer *queue.Producer
}

func NewNotificationService(producer *queue.Producer) NotificationService {
	return &notificationService{producer: producer}
}

func (s *notificationService) SendNotification(ctx context.Context, n models.Notification) error {
	if n.Channel == "" || n.To == "" || n.Body == "" {
		return fmt.Errorf("channel, to, and body are required")
	}

	return s.producer.Publish(ctx, n)
}