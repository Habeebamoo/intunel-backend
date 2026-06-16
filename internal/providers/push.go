package providers

import (
	"context"
	"fmt"
	"log"

	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"github.com/Habeebamoo/tunnl-backend/internal/models"
)

type PushProvider struct {
	fcmKey string
}

func NewPushProvider(cfg *configs.Config) *PushProvider {
	return &PushProvider{fcmKey: cfg.FCMServerKey}
}

func (p *PushProvider) Send(ctx context.Context, n models.Notification) error {
	// TODO: implement FCM call
	log.Printf("PUSH → token: %s | title: %s | body: %s", n.To, n.Title, n.Body)
	fmt.Println("Push not yet implemented")
	return nil
}