package database

import (
	"log"

	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(cfg *configs.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.PostgresUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}

	log.Println("Postgres connected")
	return db
}