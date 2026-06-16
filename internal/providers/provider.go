package providers

import (
	"context"

	"github.com/Habeebamoo/tunnl-backend/internal/models"
)

type Provider interface {
	Send(ctx context.Context, n models.Notification) error
}