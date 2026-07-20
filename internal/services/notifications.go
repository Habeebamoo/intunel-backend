package services

import (
	"context"
	"fmt"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/Habeebamoo/intunel-backend/internal/queue"
	"github.com/Habeebamoo/intunel-backend/internal/repositories"
	"github.com/Habeebamoo/intunel-backend/internal/utils"
)

type NotificationService interface {
	SendNotification(ctx context.Context, n models.Notification) error
}

type notificationService struct {
	producer    *queue.Producer
	scheduledRepo repositories.ScheduledNotificationRepository
}

func NewNotificationService(producer *queue.Producer, scheduledRepo repositories.ScheduledNotificationRepository) NotificationService {
	return &notificationService{
		producer:    producer,
		scheduledRepo: scheduledRepo,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, n models.Notification) error {
    if n.Channel == "" || n.To == "" || n.Body == "" {
        return fmt.Errorf("channel, to, and body are required")
    }

    // If date + time provided, it's a scheduled notification
    if n.Date != "" && n.Time != "" {
        scheduledAt, err := utils.ParseScheduledAt(n.Date, n.Time, n.Timezone)
        if err != nil {
            return fmt.Errorf("invalid schedule: %w", err)
        }

        scheduled := &models.ScheduledNotification{
            Channel:     n.Channel,
            To:          n.To,
            Title:       n.Title,
            Body:        n.Body,
            ScheduledAt: scheduledAt,
        }

        return s.scheduledRepo.Create(ctx, scheduled)
    }

    // Otherwise publish immediately
    return s.producer.Publish(ctx, n)
}