package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	Env               string
	RedisUrl          string
	PostgresUrl       string
	TelegramBotToken  string
	TermiiAPIKey      string
	FCMServerKey      string
}

func Load() *Config {
	// load .env file (ignore error in prod)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	return &Config{
		Port:     port,
		Env:      env,
		RedisUrl: redisURL,
		PostgresUrl:      os.Getenv("POSTGRES_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TermiiAPIKey:     os.Getenv("TERMII_API_KEY"),
		FCMServerKey:     os.Getenv("FCM_SERVER_KEY"),
	}
}