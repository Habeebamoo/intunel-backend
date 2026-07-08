package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	Name           string     `gorm:"not null" json:"name"`
	Email          string     `gorm:"uniqueIndex;not null" json:"email"`
	Avatar         string     `gorm:"default:null" json:"avatar,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.UserID = uuid.New()
	return nil
}