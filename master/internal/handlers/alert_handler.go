package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/dto"
	"github.com/sneaky-developer/UptimeHub/master/internal/models"
	"github.com/sneaky-developer/UptimeHub/master/internal/notifier"
)

type AlertHandler struct {
	db       *gorm.DB
	notifier *notifier.Service
}

func NewAlertHandler(db *gorm.DB, notifier *notifier.Service) *AlertHandler {
	return &AlertHandler{db: db, notifier: notifier}
}

// ListChannels returns all notification channels
func (h *AlertHandler) ListChannels(c *gin.Context) {
	var channels []models.NotificationChannel
	if err := h.db.Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to fetch channels"})
		return
	}
	c.JSON(http.StatusOK, channels)
}

// CreateChannel creates a new notification channel
func (h *AlertHandler) CreateChannel(c *gin.Context) {
	var req models.NotificationChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := h.db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create channel"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// UpdateChannel updates an existing channel
func (h *AlertHandler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	var channel models.NotificationChannel
	if err := h.db.First(&channel, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Channel not found"})
		return
	}

	var req models.NotificationChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request"})
		return
	}

	channel.Name = req.Name
	channel.Type = req.Type
	channel.Config = req.Config
	channel.Enabled = req.Enabled
	channel.UpdatedAt = time.Now()

	if err := h.db.Save(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to update channel"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// DeleteChannel deletes a channel
func (h *AlertHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.NotificationChannel{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to delete channel"})
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Channel deleted"})
}

// TestChannel sends a test alert to the channel
func (h *AlertHandler) TestChannel(c *gin.Context) {
	id := c.Param("id")
	var channel models.NotificationChannel
	if err := h.db.First(&channel, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Channel not found"})
		return
	}

	// Create a dummy incident for testing
	dummyIncident := models.Incident{
		ID:          uuid.New(),
		Title:       "Test Alert",
		Description: "This is a test alert from UptimeHub to verify the notification channel.",
		Status:      "investigating",
		Severity:    "minor",
		StartedAt:   time.Now(),
	}

	// We abuse the notifier internal methods here?
	// The notifier doesn't expose `SendToChannel`.
	// Ideally `notifier` should expose a usage method.
	// But `SendToChannel` is private.
	// I'll update `notifier/notifier.go` to export `SendToChannel`?
	// Or just create a temporary workaround:
	// Use a new method `TestChannel(channel, incident)` in notifier.
	
	// I'll come back to this. For now returning "Not Implemented" or just skipping logic?
	// I should update notifier to allow testing specific channel.
	
	err := h.notifier.TestChannel(channel, &dummyIncident)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Test failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Test alert sent successfully"})
}
