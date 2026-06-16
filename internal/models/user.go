package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)


type User struct {
    ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
    Name            string    
    Email           string     `gorm:"uniqueIndex;not null"`
    Avatar          string    
    Provider        string     // "google" or "github"
    ProviderID      string     // their ID on that platform
    TelegramChatID  string
    FCMToken        string
    Phone           string
    CreatedAt       time.Time
    UpdatedAt    	  time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}