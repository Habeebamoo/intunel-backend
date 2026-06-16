package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"github.com/Habeebamoo/tunnl-backend/internal/database"
	"github.com/Habeebamoo/tunnl-backend/internal/handlers"
	"github.com/Habeebamoo/tunnl-backend/internal/queue"
	"github.com/Habeebamoo/tunnl-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type App struct {
	server *http.Server
	router *gin.Engine
	config *configs.Config
}

func New() *App {
	cfg := configs.Load()

	// Redis init
	redisOpts, err := redis.ParseURL(cfg.RedisUrl)
	if err != nil {
		log.Fatalf("invalid redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)

	// Ping Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}

	log.Println("Redis connected")

	// Postgres init
	db := database.NewPostgres(cfg)
	database.Migrate(db)

	// Gin init
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()

	// Services init
	producer := queue.NewProducer(redisClient)
	notificationService := services.NewNotificationService(producer)

	// Handlers init
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Routes
	RegisterRoutes(router, notificationHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		server: server,
		router: router,
		config: cfg,
	}
}

func (a *App) Run() error {
	go func() {
		log.Printf("Tunnl running on PORT :%s", a.config.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.Shutdown(ctx)
}