package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const stateTTL = 5 * time.Minute

type OAuthStateStore struct {
	client *redis.Client
}

func NewOAuthStateStore(client *redis.Client) *OAuthStateStore {
	return &OAuthStateStore{client: client}
}

func (s *OAuthStateStore) Generate(ctx context.Context) (string, error) {
	state := uuid.New().String()
	key := fmt.Sprintf("oauth:state:%s", state)
	if err := s.client.Set(ctx, key, "1", stateTTL).Err(); err != nil {
		return "", fmt.Errorf("failed to store state: %w", err)
	}
	return state, nil
}

func (s *OAuthStateStore) Verify(ctx context.Context, state string) error {
	key := fmt.Sprintf("oauth:state:%s", state)
	val, err := s.client.GetDel(ctx, key).Result()
	if err != nil || val == "" {
		return fmt.Errorf("invalid or expired state")
	}
	return nil
}