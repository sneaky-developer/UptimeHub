package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
}

type SlackPayload struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	Channel     string       `json:"channel,omitempty"`
}

type Attachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
	Ts     int64  `json:"ts"`
}

func (s *Service) sendSlack(channel models.NotificationChannel, incident *models.Incident) error {
	var cfg SlackConfig
	if err := parseConfig(channel.Config, &cfg); err != nil {
		return fmt.Errorf("invalid slack config: %w", err)
	}

	if cfg.WebhookURL == "" {
		return fmt.Errorf("missing webhook_url")
	}

	color := "#36a64f" // green
	if incident.Status == "investigating" {
		color = "#d00000" // red
	} else if incident.Status == "identified" {
		color = "#ffcc00" // yellow
	}

	payload := SlackPayload{
		Text:    fmt.Sprintf("*%s Incident Update*", incident.Status),
		Channel: cfg.Channel,
		Attachments: []Attachment{
			{
				Color:  color,
				Title:  incident.Title,
				Text:   incident.Description,
				Footer: "UptimeHub",
				Ts:     time.Now().Unix(),
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(cfg.WebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}
