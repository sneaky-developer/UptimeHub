package services

import (
	"log"
	"math"
	"time"

	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

// AggregatorService computes hourly/daily uptime aggregations and prunes
// old raw check results
type AggregatorService struct {
	db            *gorm.DB
	retentionDays int
}

// NewAggregatorService creates a new AggregatorService. retentionDays controls
// how long raw check results are kept (aggregations are kept forever).
func NewAggregatorService(db *gorm.DB, retentionDays int) *AggregatorService {
	if retentionDays < 1 {
		retentionDays = 90
	}
	return &AggregatorService{db: db, retentionDays: retentionDays}
}

// Start begins the periodic aggregation workers
func (s *AggregatorService) Start() {
	go s.runHourlyAggregation()
	go s.runDailyAggregation()
	go s.runRetention()
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
	// Run immediately, then hourly so today's partial stats stay fresh —
	// without this a new install shows an empty status page until midnight
	s.aggregate("daily", 24*time.Hour)
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.aggregate("daily", 24*time.Hour)
	}
}

// runRetention prunes raw check results past the retention window once a day
func (s *AggregatorService) runRetention() {
	s.pruneCheckResults()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.pruneCheckResults()
	}
}

func (s *AggregatorService) pruneCheckResults() {
	cutoff := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	result := s.db.Where("checked_at < ?", cutoff).Delete(&models.CheckResult{})
	if result.Error != nil {
		log.Printf("⚠️  Check result retention failed: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("🧹 Pruned %d check results older than %d days", result.RowsAffected, s.retentionDays)
	}
}

func (s *AggregatorService) aggregate(periodType string, period time.Duration) {
	// Aggregate the previous complete period and the current partial one, so
	// the status page reflects data as soon as checks start flowing
	now := time.Now().UTC()
	var currentStart time.Time

	switch periodType {
	case "hourly":
		currentStart = now.Truncate(time.Hour)
	case "daily":
		currentStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}

	s.aggregatePeriod(periodType, currentStart.Add(-period), currentStart)
	s.aggregatePeriod(periodType, currentStart, currentStart.Add(period))
}

func (s *AggregatorService) aggregatePeriod(periodType string, start, end time.Time) {
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
}
