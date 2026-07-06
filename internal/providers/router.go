package providers

import (
	"context"
	"fmt"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/models"
)

type Router struct {
	providers map[string]Provider
}

func NewRouter(cfg *configs.Config) *Router {
	return &Router{
		providers: map[string]Provider{
			"email": NewEmailProvider(cfg),
		},
	}
}

func (r *Router) Route(ctx context.Context, n models.Notification) error {
	p, ok := r.providers[n.Channel]
	if !ok {
		return fmt.Errorf("unknown channel: %s", n.Channel)
	}
	return p.Send(ctx, n)
}