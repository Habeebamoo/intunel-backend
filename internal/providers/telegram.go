package providers

import (
	"context"
	"fmt"
	"log"

	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"github.com/Habeebamoo/tunnl-backend/internal/models"
)

type TelegramProvider struct {
	botToken string
}

func NewTelegramProvider(cfg *configs.Config) *TelegramProvider {
	return &TelegramProvider{botToken: cfg.TelegramBotToken}
}

func (t *TelegramProvider) Send(ctx context.Context, n models.Notification) error {
	// TODO: implement Telegram Bot API call
	log.Printf("TELEGRAM → chat_id: %s | message: %s", n.To, n.Body)
	fmt.Println("Telegram not yet implemented")
	return nil
}