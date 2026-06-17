package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Habeebamoo/tunnl-backend/internal/configs"
	"github.com/Habeebamoo/tunnl-backend/internal/models"
)

type SMSProvider struct {
	apiKey   string
	senderID string
	client   *http.Client
}

func NewSMSProvider(cfg *configs.Config) *SMSProvider {
	return &SMSProvider{
		apiKey:   cfg.TermiiAPIKey,
		senderID: cfg.TermiiSenderID,
		client:   &http.Client{},
	}
}

func (s *SMSProvider) Send(ctx context.Context, n models.Notification) error {
	to := strings.TrimPrefix(n.To, "+")

	payload := map[string]interface{}{
		"api_key": s.apiKey,
		"to":      to,
		"from":    s.senderID,
		"sms":     fmt.Sprintf("%s\n%s", n.Title, n.Body),
		"type":    "plain",
		"channel": "generic",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://v3.api.termii.com/api/sms/send", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("termii request failed: %w", err)
	}
	defer resp.Body.Close()

	respb, _ := io.ReadAll(resp.Body)
	log.Printf("Termii response: %s", string(respb))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("termii returned status: %d", resp.StatusCode)
	}

	return nil
}