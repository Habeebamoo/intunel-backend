package providers

import (
	"context"
	"fmt"
	"log"

	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"github.com/Habeebamoo/tunnl-backend/internal/models"
)

type SMSProvider struct {
	apiKey string
}

func NewSMSProvider(cfg *configs.Config) *SMSProvider {
	return &SMSProvider{apiKey: cfg.TermiiAPIKey}
}

func (s *SMSProvider) Send(ctx context.Context, n models.Notification) error {
	// TODO: implement Termii API call
	log.Printf("SMS → phone: %s | message: %s", n.To, n.Body)
	fmt.Println("SMS not yet implemented")
	return nil
}