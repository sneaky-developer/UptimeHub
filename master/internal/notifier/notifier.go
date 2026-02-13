package notifier

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
	"gorm.io/gorm"
)

// Service manages notification channels and sends alerts
type Service struct {
	db *gorm.DB
}

// NewService creates a new notification service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Notify sends alerts for a given incident to all enabled channels
func (s *Service) Notify(incident *models.Incident) {
	var channels []models.NotificationChannel
	if err := s.db.Where("enabled = ?", true).Find(&channels).Error; err != nil {
		log.Printf("Failed to fetch notification channels: %v", err)
		return
	}

	if len(channels) == 0 {
		return
	}

	log.Printf("📢 Sending alerts for incident %s (%s) to %d channels", incident.Title, incident.Status, len(channels))

	for _, channel := range channels {
		go s.sendToChannel(channel, incident)
	}
}

func (s *Service) sendToChannel(channel models.NotificationChannel, incident *models.Incident) {
	var err error
	
	switch channel.Type {
	case "email":
		err = s.sendEmail(channel, incident)
	case "slack":
		err = s.sendSlack(channel, incident)
	case "webhook":
		err = s.sendWebhook(channel, incident)
	default:
		err = fmt.Errorf("unknown channel type: %s", channel.Type)
	}

	status := "sent"
	errorMsg := ""
	if err != nil {
		status = "failed"
		errorMsg = err.Error()
		log.Printf("❌ Failed to alert channel %s: %v", channel.Name, err)
	} else {
		log.Printf("✅ Alert sent to %s", channel.Name)
	}

	// Log the attempt
	alertLog := models.AlertLog{
		ChannelID:  channel.ID,
		IncidentID: incident.ID,
		Status:     status,
		Error:      errorMsg,
		CreatedAt:  time.Now(),
	}
	s.db.Create(&alertLog)
}

// TestChannel sends a test alert to a specific channel
func (s *Service) TestChannel(channel models.NotificationChannel, incident *models.Incident) error {
	// Attempt to send
	s.sendToChannel(channel, incident)
	// sendToChannel handles logging, but doesn't return error directly (it swallows it)
	// I should refactor sendToChannel to return error?
	// Or just check the log.
	
	// Let's refactor sendToChannel.
	// Oh wait, sendToChannel is async in Notify but sync here?
	// I'll refactor sendToChannel to return error and log inside a wrapper.
	
	switch channel.Type {
	case "email":
		return s.sendEmail(channel, incident)
	case "slack":
		return s.sendSlack(channel, incident)
	case "webhook":
		return s.sendWebhook(channel, incident)
	default:
		return fmt.Errorf("unknown channel type: %s", channel.Type)
	}
}

// Helper to parse config
func parseConfig(config models.JSON, target interface{}) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}
