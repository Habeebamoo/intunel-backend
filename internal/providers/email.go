package providers

import (
	"context"
	"fmt"
	"log"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/models"
	"github.com/resend/resend-go/v3"
)

type EmailProvider struct {
	ResendApiKey string
}

func NewEmailProvider(cfg *configs.Config) *EmailProvider {
	return &EmailProvider{ResendApiKey: cfg.ResendApiKey}
}

func (e *EmailProvider) Send(ctx context.Context, n models.Notification) error {
	client := resend.NewClient(e.ResendApiKey)
	title := fmt.Sprintf("%s <hello@myclivo.com>", n.Title)

	params := &resend.SendEmailRequest{
		From: title,
		To: []string{n.To},
		Subject: n.Title,
		Html: n.Body,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return err
	}

	log.Printf("Email sent: %v", n)
	return nil
}