package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	UserID    string  `json:"user_id"`
	Channel   string  `json:"channel"`
	To        string  `json:"to"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Date      string  `json:"date,omitempty"`
	Time      string  `json:"time,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
}

type ScheduledNotification struct {
	UserID       uuid.UUID   `gorm:"type:uuid;primaryKey" json:"user_id"`
	Channel      string      `gorm:"not null" json:"channel"`
	To           string      `gorm:"column:to_address;not null" json:"to"`
	Title        string      `gorm:"default:null" json:"title"`
	Body         string      `gorm:"not null" json:"body"`
	ScheduledAt  time.Time   `gorm:"not null" json:"scheduled_at"`
	Status       string      `gorm:"default:scheduled" json:"status"`
	PublishedAt  *time.Time  `gorm:"default:null" json:"published_at"`
	CreatedAt    time.Time   `json:"created_at"`
}

func (s *ScheduledNotification) BeforeCreate(tx *gorm.DB) error {
	s.UserID = uuid.New()
	return nil
}
