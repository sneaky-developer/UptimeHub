package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds agent configuration
type Config struct {
	// Master server connection
	MasterURL string
	AgentName string
	EnrollmentToken string

	// Agent token (obtained after registration, stored in memory)
	AgentToken string

	// Check defaults
	DefaultInterval    time.Duration
	DefaultTimeout     time.Duration
	DefaultRetries     int
	FailureThreshold   int

	// Heartbeat
	HeartbeatInterval  time.Duration
	ConfigSyncInterval time.Duration

	// Kubernetes
	KubeNamespace string // empty = all namespaces
	InCluster     bool
}

// Load reads agent config from environment variables
func Load() *Config {
	return &Config{
		MasterURL:          getEnv("MASTER_URL", "http://localhost:8080"),
		AgentName:          getEnv("AGENT_NAME", "default-agent"),
		EnrollmentToken:    getEnv("ENROLLMENT_TOKEN", ""),
		AgentToken:         getEnv("AGENT_TOKEN", ""),
		DefaultInterval:    getDurationEnv("DEFAULT_INTERVAL", 30*time.Second),
		DefaultTimeout:     getDurationEnv("DEFAULT_TIMEOUT", 10*time.Second),
		DefaultRetries:     getEnvInt("DEFAULT_RETRIES", 3),
		FailureThreshold:   getEnvInt("FAILURE_THRESHOLD", 3),
		HeartbeatInterval:  getDurationEnv("HEARTBEAT_INTERVAL", 60*time.Second),
		ConfigSyncInterval: getDurationEnv("CONFIG_SYNC_INTERVAL", 60*time.Second),
		KubeNamespace:      getEnv("KUBE_NAMESPACE", ""),
		InCluster:          getEnvBool("IN_CLUSTER", true),
	}
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

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if secs, err := strconv.Atoi(value); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return fallback
}
