package providers

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"github.com/Habeebamoo/tunnl-backend/internal/models"
	"google.golang.org/api/option"
)

type PushProvider struct {
	client *messaging.Client
}

func NewPushProvider(cfg *configs.Config) *PushProvider {
	opt := option.WithCredentialsFile(cfg.FirebaseCredentialsPath)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("failed to initialize firebase app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("failed to initialize firebase messaging: %v", err)
	}

	return &PushProvider{client: client}
}

func (p *PushProvider) Send(ctx context.Context, n models.Notification) error {
	message := &messaging.Message{
		Token: n.To,
		Notification: &messaging.Notification{
			Title: n.Title,
			Body:  n.Body,
		},
		Data: map[string]string{
			"user_id": n.UserID,
			"channel": n.Channel,
		},
	}

	response, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	log.Printf("FCM response: %s", response)
	return nil
}