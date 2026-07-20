package worker

import (
	"context"
	"log"
	"time"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/Habeebamoo/intunel-backend/internal/queue"
	"github.com/Habeebamoo/intunel-backend/internal/repositories"
)

type Scheduler struct {
	repo      repositories.ScheduledNotificationRepository
	producer  *queue.Producer
}

func NewScheduler(repo repositories.ScheduledNotificationRepository, producer *queue.Producer) *Scheduler {
	return &Scheduler{repo: repo, producer: producer}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Scheduler: running...")

	for {
		select {
		case <-ticker.C:
			s.publishDue(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) publishDue(ctx context.Context) {
	log.Printf("Scheduler: checking for new notifs")
	
	notifications, err := s.repo.FindDue(ctx)
	if err != nil {
		log.Printf("Scheduler: error fetching due notifications: %v\n", err)
		return
	}

	if len(notifications) == 0 {
		return
	}

	log.Printf("Scheduler: found %d due notifications\n", len(notifications))

	for _, n := range notifications {
		notification := models.Notification{
			Channel: n.Channel,
			To:      n.To,
			Title:   n.Title,
			Body:    n.Body,
		}

		if err := s.producer.Publish(ctx, notification); err != nil {
			log.Printf("Scheduler: failed to publish notification %s: %v\n", n.UserID, err)
			continue
		}

		if err := s.repo.MarkQueued(ctx, n.UserID.String()); err != nil {
			log.Printf("Scheduler: failed to mark notification %s as queued: %v\n", n.UserID, err)
		}

		log.Printf("Scheduler: published notification %s → [%s] %s\n", n.UserID, n.Channel, n.To)
	}
}