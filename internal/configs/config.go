package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	Env                      string
	RedisUrl                 string
	PostgresUrl              string
	TermiiAPIKey             string
	TermiiSenderID           string
	FCMServerKey             string
	GoogleClientID           string
	GoogleClientSecret       string
	GitHubClientID           string
	GitHubClientSecret       string
	FirebaseCredentialsJSON  string
	FrontendUrl              string
	JwtSecret                string
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
		Port:                     port,
		Env:                      env,
		RedisUrl: 					      redisURL,
		PostgresUrl:              os.Getenv("POSTGRES_URL"),
		TermiiAPIKey:             os.Getenv("TERMII_API_KEY"),
		TermiiSenderID:           os.Getenv("TERMII_SENDER_ID"),
		FCMServerKey:             os.Getenv("FCM_SERVER_KEY"),
		GoogleClientID:           os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:       os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:           os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:       os.Getenv("GITHUB_CLIENT_SECRET"),
		FirebaseCredentialsJSON:  os.Getenv("FIREBASE_CREDENTIALS_JSON"),
		FrontendUrl:              os.Getenv("FRONTEND_URL"),
		JwtSecret:                os.Getenv("JWT_SECRET"),
	}
}