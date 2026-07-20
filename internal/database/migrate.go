package database

import (
	"log"

	"github.com/Habeebamoo/intunel-backend/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
    &models.User{},
    &models.ScheduledNotification{},
	)

	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("Database migrated")
}