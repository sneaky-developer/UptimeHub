package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	AppEnv        string
	AppPort       string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTSecret     string
	CORSOrigins   string
	AdminEmail    string
	AdminPassword string

	// CheckRetentionDays is how long raw check results are kept
	CheckRetentionDays int
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("APP_PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "uptimehub"),
		DBPassword:  getEnv("DB_PASSWORD", "uptimehub_secret"),
		DBName:      getEnv("DB_NAME", "uptimehub"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		CORSOrigins:   getEnv("CORS_ORIGINS", "http://localhost:3000"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@uptimehub.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),

		CheckRetentionDays: getEnvInt("CHECK_RETENTION_DAYS", 90),
	}
}

// insecureJWTSecrets are placeholder values that must never reach production
var insecureJWTSecrets = map[string]bool{
	"":                                   true,
	"change-me-in-production":            true,
	"dev-jwt-secret-change-in-production": true,
}

// Validate rejects insecure settings when not running in development
func (c *Config) Validate() error {
	if c.IsDevelopment() {
		return nil
	}
	if insecureJWTSecrets[c.JWTSecret] {
		return fmt.Errorf("JWT_SECRET must be set to a strong random value when APP_ENV is not development")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters when APP_ENV is not development")
	}
	return nil
}

// DSN returns the PostgreSQL connection string
func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=" + c.DBSSLMode +
		" TimeZone=UTC"
}

// IsDevelopment returns true if running in dev mode
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
