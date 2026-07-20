package repositories

import (
	"context"
	"time"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"gorm.io/gorm"
)

type ScheduledNotificationRepository interface {
	Create(ctx context.Context, n *models.ScheduledNotification) error
	FindDue(ctx context.Context) ([]models.ScheduledNotification, error)
	MarkQueued(ctx context.Context, id string) error
}

type scheduledNotificationRepository struct {
	db *gorm.DB
}

func NewScheduledNotificationRepository(db *gorm.DB) ScheduledNotificationRepository {
	return &scheduledNotificationRepository{db: db}
}

func (r *scheduledNotificationRepository) Create(ctx context.Context, n *models.ScheduledNotification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *scheduledNotificationRepository) FindDue(ctx context.Context) ([]models.ScheduledNotification, error) {
	var notifications []models.ScheduledNotification

	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", "scheduled", time.Now().UTC()).
		Find(&notifications).Error

	return notifications, err
}

func (r *scheduledNotificationRepository) MarkQueued(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.ScheduledNotification{}).
		Where("user_id = ?", id).
		Updates(map[string]interface{}{
			"status":       "queued",
			"published_at": now,
		}).Error
}