package database

import (
	"fmt"
	"log"
	"time"

	"github.com/sneaky-developer/UptimeHub/master/internal/config"
	"github.com/sneaky-developer/UptimeHub/master/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL and runs auto-migrations
func Connect(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run auto-migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database connected and migrations applied")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Agent{},
		&models.Service{},
		&models.CheckResult{},
		&models.UptimeAggregation{},
		&models.Incident{},
		&models.IncidentUpdate{},
		&models.MaintenanceWindow{},
		&models.AdminUser{},
		&models.NotificationChannel{},
		&models.AlertLog{},
	)
}
