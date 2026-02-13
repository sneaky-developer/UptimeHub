package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Agent represents a registered worker agent running in a Kubernetes cluster
type Agent struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	ClusterName   string         `gorm:"size:255;not null" json:"cluster_name"`
	Token         string         `gorm:"size:512;not null;uniqueIndex" json:"-"`
	Status        string         `gorm:"size:20;default:'pending'" json:"status"` // pending, active, inactive
	LastHeartbeat *time.Time     `json:"last_heartbeat"`
	Metadata      JSON           `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Services      []Service      `gorm:"foreignKey:AgentID" json:"services,omitempty"`
}

// Service represents a monitored endpoint
type Service struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AgentID          *uuid.UUID     `gorm:"type:uuid;index" json:"agent_id"`
	Name             string         `gorm:"size:255;not null" json:"name"`
	URL              string         `gorm:"size:2048;not null" json:"url"`
	CheckInterval    int            `gorm:"default:30" json:"check_interval"`     // seconds
	Timeout          int            `gorm:"default:10" json:"timeout"`            // seconds
	Retries          int            `gorm:"default:3" json:"retries"`
	FailureThreshold int            `gorm:"default:3" json:"failure_threshold"`
	Status           string         `gorm:"size:20;default:'unknown'" json:"status"` // up, down, degraded, unknown
	IsPublic         bool           `gorm:"default:true" json:"is_public"`
	GroupName        string         `gorm:"size:255" json:"group_name"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Agent            *Agent         `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

// CheckResult stores an individual health check result
type CheckResult struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceID    uuid.UUID `gorm:"type:uuid;not null;index:idx_check_service_time" json:"service_id"`
	AgentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"agent_id"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time_ms"` // milliseconds
	IsUp         bool      `gorm:"not null" json:"is_up"`
	ErrorMessage string    `gorm:"type:text" json:"error_message,omitempty"`
	CheckedAt    time.Time `gorm:"not null;index:idx_check_service_time" json:"checked_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// UptimeAggregation stores hourly/daily uptime stats
type UptimeAggregation struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_agg_unique" json:"service_id"`
	PeriodStart   time.Time `gorm:"not null;uniqueIndex:idx_agg_unique" json:"period_start"`
	PeriodEnd     time.Time `gorm:"not null" json:"period_end"`
	PeriodType    string    `gorm:"size:10;not null;uniqueIndex:idx_agg_unique" json:"period_type"` // hourly, daily
	TotalChecks   int       `gorm:"default:0" json:"total_checks"`
	Successful    int       `gorm:"default:0" json:"successful"`
	UptimePct     float64   `gorm:"type:decimal(5,2)" json:"uptime_pct"`
	AvgResponseMs int       `json:"avg_response_ms"`
}

// Incident represents a service incident (auto or manual)
type Incident struct {
	ID         uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceID  *uuid.UUID       `gorm:"type:uuid;index" json:"service_id"`
	Title      string           `gorm:"size:500;not null" json:"title"`
	Description string          `gorm:"type:text" json:"description"`
	Status     string           `gorm:"size:20;default:'investigating'" json:"status"` // investigating, identified, monitoring, resolved
	Severity   string           `gorm:"size:20;default:'minor'" json:"severity"`       // minor, major, critical
	StartedAt  time.Time        `gorm:"default:now()" json:"started_at"`
	ResolvedAt *time.Time       `json:"resolved_at"`
	IsManual   bool             `gorm:"default:false" json:"is_manual"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Updates    []IncidentUpdate `gorm:"foreignKey:IncidentID" json:"updates,omitempty"`
	Service    *Service         `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

// IncidentUpdate represents a status update on an incident
type IncidentUpdate struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IncidentID uuid.UUID `gorm:"type:uuid;not null;index" json:"incident_id"`
	Status     string    `gorm:"size:20;not null" json:"status"`
	Message    string    `gorm:"type:text;not null" json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

// MaintenanceWindow represents a scheduled maintenance period
type MaintenanceWindow struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title          string     `gorm:"size:500;not null" json:"title"`
	Description    string     `gorm:"type:text" json:"description"`
	ServiceIDs     UUIDArray  `gorm:"type:uuid[]" json:"service_ids"`
	ScheduledStart time.Time  `gorm:"not null" json:"scheduled_start"`
	ScheduledEnd   time.Time  `gorm:"not null" json:"scheduled_end"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AdminUser represents an admin dashboard user
type AdminUser struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Name         string         `gorm:"size:255" json:"name"`
	Role         string         `gorm:"size:20;default:'admin'" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// NotificationChannel represents a destination for alerts (email, slack, etc.)
type NotificationChannel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Type      string         `gorm:"size:50;not null" json:"type"` // email, slack, webhook
	Config    JSON           `gorm:"type:jsonb;default:'{}'" json:"config"`
	Enabled   bool           `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AlertLog tracks sent alerts
type AlertLog struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ChannelID   uuid.UUID           `gorm:"type:uuid;not null;index" json:"channel_id"`
	IncidentID  uuid.UUID           `gorm:"type:uuid;not null;index" json:"incident_id"`
	Status      string              `gorm:"size:50;not null" json:"status"` // sent, failed
	Error       string              `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	Channel     *NotificationChannel `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Incident    *Incident           `gorm:"foreignKey:IncidentID" json:"incident,omitempty"`
}
