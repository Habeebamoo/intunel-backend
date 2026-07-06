package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/providers"
	"github.com/Habeebamoo/intunel-backend/internal/queue"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := configs.Load()

	redisOpts, err := redis.ParseURL(cfg.RedisUrl)
	if err != nil {
		log.Fatalf("invalid redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ping Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}
	log.Println("Worker: Redis connected")

	// Init providers
	router := providers.NewRouter(cfg)

	// Init consumer
	consumer := queue.NewConsumer(redisClient, router)

	// Start consuming
	go consumer.Start(ctx)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")
	cancel()
}