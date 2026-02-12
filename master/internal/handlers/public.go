package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/dto"
	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

// PublicHandler handles public status page API endpoints (no auth required)
type PublicHandler struct {
	db *gorm.DB
}

// NewPublicHandler creates a new PublicHandler
func NewPublicHandler(db *gorm.DB) *PublicHandler {
	return &PublicHandler{db: db}
}

// Status handles GET /api/status — returns current status of all public services
func (h *PublicHandler) Status(c *gin.Context) {
	var services []models.Service
	h.db.Where("is_public = true").Order("group_name, name").Find(&services)

	result := make([]dto.PublicServiceStatus, 0, len(services))
	for _, svc := range services {
		// Calculate overall uptime from daily aggregations (last 90 days)
		var avgUptime float64
		h.db.Model(&models.UptimeAggregation{}).
			Where("service_id = ? AND period_type = 'daily' AND period_start > ?", svc.ID, time.Now().AddDate(0, 0, -90)).
			Select("COALESCE(AVG(uptime_pct), 100)").
			Row().Scan(&avgUptime)

		// Get last 90 days of daily history
		history := h.getDailyHistory(svc.ID, 90)

		result = append(result, dto.PublicServiceStatus{
			ID:        svc.ID,
			Name:      svc.Name,
			Status:    svc.Status,
			GroupName: svc.GroupName,
			UptimePct: avgUptime,
			History:   history,
		})
	}

	c.JSON(http.StatusOK, result)
}

// ServiceHistory handles GET /api/status/:id/history
func (h *PublicHandler) ServiceHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.Where("id = ? AND is_public = true", id).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Service not found"})
		return
	}

	history := h.getDailyHistory(id, 90)

	c.JSON(http.StatusOK, gin.H{
		"service": dto.PublicServiceStatus{
			ID:        service.ID,
			Name:      service.Name,
			Status:    service.Status,
			GroupName: service.GroupName,
		},
		"history": history,
	})
}

// Incidents handles GET /api/incidents — returns recent public incidents
func (h *PublicHandler) Incidents(c *gin.Context) {
	var incidents []models.Incident
	h.db.Preload("Updates", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Preload("Service").
		Where("started_at > ?", time.Now().AddDate(0, 0, -90)).
		Order("started_at DESC").
		Limit(50).
		Find(&incidents)

	c.JSON(http.StatusOK, incidents)
}

// Maintenance handles GET /api/maintenance — returns upcoming and active maintenance
func (h *PublicHandler) Maintenance(c *gin.Context) {
	var windows []models.MaintenanceWindow
	h.db.Where("scheduled_end > ?", time.Now()).
		Order("scheduled_start ASC").
		Find(&windows)

	c.JSON(http.StatusOK, windows)
}

// getDailyHistory returns daily uptime data for a service
func (h *PublicHandler) getDailyHistory(serviceID uuid.UUID, days int) []dto.UptimeDayStatus {
	since := time.Now().AddDate(0, 0, -days)

	var aggregations []models.UptimeAggregation
	h.db.Where("service_id = ? AND period_type = 'daily' AND period_start > ?", serviceID, since).
		Order("period_start ASC").
		Find(&aggregations)

	history := make([]dto.UptimeDayStatus, 0, len(aggregations))
	for _, agg := range aggregations {
		status := "up"
		if agg.UptimePct < 99.0 {
			status = "partial"
		}
		if agg.UptimePct < 95.0 {
			status = "degraded"
		}
		if agg.UptimePct < 50.0 {
			status = "down"
		}

		history = append(history, dto.UptimeDayStatus{
			Date:      agg.PeriodStart.Format("2006-01-02"),
			UptimePct: agg.UptimePct,
			Status:    status,
		})
	}

	return history
}
