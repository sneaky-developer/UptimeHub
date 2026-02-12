package services

import (
	"log"
	"math"
	"time"

	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

// AggregatorService computes hourly/daily uptime aggregations
type AggregatorService struct {
	db *gorm.DB
}

// NewAggregatorService creates a new AggregatorService
func NewAggregatorService(db *gorm.DB) *AggregatorService {
	return &AggregatorService{db: db}
}

// Start begins the periodic aggregation workers
func (s *AggregatorService) Start() {
	go s.runHourlyAggregation()
	go s.runDailyAggregation()
	log.Println("📊 Aggregation workers started")
}

func (s *AggregatorService) runHourlyAggregation() {
	// Run immediately, then every hour
	s.aggregate("hourly", time.Hour)
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.aggregate("hourly", time.Hour)
	}
}

func (s *AggregatorService) runDailyAggregation() {
	// Run immediately, then every 24 hours
	s.aggregate("daily", 24*time.Hour)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.aggregate("daily", 24*time.Hour)
	}
}

func (s *AggregatorService) aggregate(periodType string, period time.Duration) {
	// Calculate period boundaries
	now := time.Now().UTC()
	var start, end time.Time

	switch periodType {
	case "hourly":
		end = now.Truncate(time.Hour)
		start = end.Add(-time.Hour)
	case "daily":
		end = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		start = end.Add(-24 * time.Hour)
	}

	// Get all services
	var services []models.Service
	s.db.Find(&services)

	for _, svc := range services {
		var totalChecks int64
		var successfulChecks int64
		var avgResponseTime float64

		s.db.Model(&models.CheckResult{}).
			Where("service_id = ? AND checked_at >= ? AND checked_at < ?", svc.ID, start, end).
			Count(&totalChecks)

		if totalChecks == 0 {
			continue
		}

		s.db.Model(&models.CheckResult{}).
			Where("service_id = ? AND checked_at >= ? AND checked_at < ? AND is_up = true", svc.ID, start, end).
			Count(&successfulChecks)

		s.db.Model(&models.CheckResult{}).
			Where("service_id = ? AND checked_at >= ? AND checked_at < ?", svc.ID, start, end).
			Select("COALESCE(AVG(response_time), 0)").
			Row().Scan(&avgResponseTime)

		uptimePct := float64(successfulChecks) / float64(totalChecks) * 100
		uptimePct = math.Round(uptimePct*100) / 100

		agg := models.UptimeAggregation{
			ServiceID:     svc.ID,
			PeriodStart:   start,
			PeriodEnd:     end,
			PeriodType:    periodType,
			TotalChecks:   int(totalChecks),
			Successful:    int(successfulChecks),
			UptimePct:     uptimePct,
			AvgResponseMs: int(avgResponseTime),
		}

		// Upsert: insert or update if exists
		s.db.Where(
			"service_id = ? AND period_start = ? AND period_type = ?",
			svc.ID, start, periodType,
		).Assign(agg).FirstOrCreate(&agg)
	}

	log.Printf("📊 %s aggregation completed for %s", periodType, start.Format("2006-01-02 15:04"))
}
