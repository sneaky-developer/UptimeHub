package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
	"github.com/sneaky-developer/UptimeHub/master/internal/notifier"
)

// MonitorService handles business logic for check results and service status
type MonitorService struct {
	db       *gorm.DB
	notifier *notifier.Service
}

// NewMonitorService creates a new MonitorService
func NewMonitorService(db *gorm.DB, notifier *notifier.Service) *MonitorService {
	return &MonitorService{db: db, notifier: notifier}
}

// ProcessCheckResults processes a batch of check results from an agent
func (s *MonitorService) ProcessCheckResults(agentID uuid.UUID, results []models.CheckResult) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	// Bulk insert check results
	for i := range results {
		results[i].AgentID = agentID
	}

	if err := s.db.Create(&results).Error; err != nil {
		return 0, fmt.Errorf("failed to insert check results: %w", err)
	}

	// Update each affected service's status once, sequentially in a single
	// goroutine — concurrent updates for the same service can race and
	// double-create incidents
	seen := make(map[uuid.UUID]bool, len(results))
	serviceIDs := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		if !seen[result.ServiceID] {
			seen[result.ServiceID] = true
			serviceIDs = append(serviceIDs, result.ServiceID)
		}
	}

	go func() {
		for _, id := range serviceIDs {
			s.updateServiceStatus(id)
		}
	}()

	return len(results), nil
}

// updateServiceStatus recalculates a service's status based on recent checks
func (s *MonitorService) updateServiceStatus(serviceID uuid.UUID) {
	var service models.Service
	if err := s.db.First(&service, serviceID).Error; err != nil {
		log.Printf("Error finding service %s: %v", serviceID, err)
		return
	}

	// Get the last N checks based on failure_threshold
	threshold := service.FailureThreshold
	if threshold < 1 {
		threshold = 3
	}

	var recentChecks []models.CheckResult
	s.db.Where("service_id = ?", serviceID).
		Order("checked_at DESC").
		Limit(threshold).
		Find(&recentChecks)

	if len(recentChecks) == 0 {
		return
	}

	oldStatus := service.Status
	newStatus := computeStatus(recentChecks, threshold)

	if oldStatus != newStatus {
		s.db.Model(&service).Update("status", newStatus)

		// Auto-incident management
		if newStatus == "down" && oldStatus != "down" {
			s.createAutoIncident(&service)
		} else if newStatus == "up" && oldStatus == "down" {
			s.resolveAutoIncidents(&service)
		}
	}
}

// computeStatus derives a service status from its most recent checks
// (ordered newest first): "down" after threshold consecutive failures,
// "degraded" while failures accumulate, otherwise "up"
func computeStatus(recentChecks []models.CheckResult, threshold int) string {
	failCount := 0
	for _, check := range recentChecks {
		if !check.IsUp {
			failCount++
		} else {
			break
		}
	}

	switch {
	case failCount >= threshold:
		return "down"
	case failCount > 0:
		return "degraded"
	default:
		return "up"
	}
}

// createAutoIncident automatically creates an incident when a service goes down
func (s *MonitorService) createAutoIncident(service *models.Service) {
	incident := models.Incident{
		ServiceID:   &service.ID,
		Title:       fmt.Sprintf("%s is experiencing issues", service.Name),
		Description: fmt.Sprintf("Automated incident: %s (%s) is not responding", service.Name, service.URL),
		Status:      "investigating",
		Severity:    "major",
		StartedAt:   time.Now(),
		IsManual:    false,
	}

	if err := s.db.Create(&incident).Error; err != nil {
		log.Printf("Error creating auto-incident for service %s: %v", service.ID, err)
		return
	}

	// Load service details for notification context
	incident.Service = service

	// Create initial update
	update := models.IncidentUpdate{
		IncidentID: incident.ID,
		Status:     "investigating",
		Message:    fmt.Sprintf("Service %s was detected as down. Investigating the issue.", service.Name),
	}
	s.db.Create(&update)

	log.Printf("🔴 Auto-incident created for service %s: %s", service.Name, incident.ID)

	// Trigger Alert
	go s.notifier.Notify(&incident)
}

// resolveAutoIncidents resolves any open auto-incidents for a recovered service
func (s *MonitorService) resolveAutoIncidents(service *models.Service) {
	now := time.Now()

	var incidents []models.Incident
	s.db.Where("service_id = ? AND is_manual = false AND resolved_at IS NULL", service.ID).Find(&incidents)

	for _, incident := range incidents {
		s.db.Model(&incident).Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
		})

		// Load service details for notification context
		incident.Service = service

		update := models.IncidentUpdate{
			IncidentID: incident.ID,
			Status:     "resolved",
			Message:    fmt.Sprintf("Service %s has recovered and is operational.", service.Name),
		}
		s.db.Create(&update)

		log.Printf("🟢 Auto-incident resolved for service %s: %s", service.Name, incident.ID)

		// Trigger Alert
		go s.notifier.Notify(&incident)
	}
}
