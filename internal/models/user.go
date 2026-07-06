package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name            string     `gorm:"not null" json:"name"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Avatar          string     `gorm:"default:null" json:"avatar,omitempty"`
	Provider        string     `gorm:"not null" json:"provider"`
	ProviderID      string     `gorm:"not null" json:"provider_id"`
	Phone           string     `gorm:"default:null" json:"phone,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}