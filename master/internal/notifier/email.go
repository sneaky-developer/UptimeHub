package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

type EmailConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
}

func (s *Service) sendEmail(channel models.NotificationChannel, incident *models.Incident) error {
	var cfg EmailConfig
	if err := parseConfig(channel.Config, &cfg); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}

	if cfg.Host == "" || cfg.Port == 0 || cfg.From == "" || len(cfg.To) == 0 {
		return fmt.Errorf("missing required email config fields")
	}

	// Message
	subject := fmt.Sprintf("[%s] %s: %s", incident.Severity, incident.Status, incident.Title)
	if incident.Service != nil {
		subject = fmt.Sprintf("[%s] %s: %s (%s)", incident.Severity, incident.Status, incident.Service.Name, incident.Title)
	}

	body := fmt.Sprintf("Incident: %s\nStatus: %s\nSeverity: %s\nStarted At: %s\n\n%s",
		incident.Title, incident.Status, incident.Severity, incident.StartedAt.Format(time.RFC1123), incident.Description)

	msg := []byte("From: " + cfg.From + "\r\n" +
		"To: " + strings.Join(cfg.To, ", ") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	// Auth
	var auth smtp.Auth
	if cfg.User != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}

	// Send
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	
	// Skip TLS verification for development/testing if needed, or use proper TLS
	// For simplicity using standard SendMail which handles StartTLS
	// But in production might need more specific TLS config
	return smtp.SendMail(addr, auth, cfg.From, cfg.To, msg)
}
