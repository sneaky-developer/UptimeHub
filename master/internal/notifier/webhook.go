package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func (s *Service) sendWebhook(channel models.NotificationChannel, incident *models.Incident) error {
	var cfg WebhookConfig
	if err := parseConfig(channel.Config, &cfg); err != nil {
		return fmt.Errorf("invalid webhook config: %w", err)
	}

	if cfg.URL == "" {
		return fmt.Errorf("missing webhook url")
	}

	// Send the incident as JSON payload
	payloadBytes, err := json.Marshal(incident)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", cfg.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UptimeHub-Notifier")
	req.Header.Set("X-UptimeHub-Event", incident.Status)

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}
