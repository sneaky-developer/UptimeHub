package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Agent DTOs ─────────────────────────────────────────────────────

type AgentRegisterRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

type AgentRegisterResponse struct {
	AgentID uuid.UUID `json:"agent_id"`
	Token   string    `json:"token"`
}

type CheckResultItem struct {
	ServiceID    uuid.UUID `json:"service_id" binding:"required"`
	URL          string    `json:"url"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time_ms"`
	IsUp         bool      `json:"is_up"`
	ErrorMessage string    `json:"error_message"`
	CheckedAt    time.Time `json:"checked_at" binding:"required"`
}

type AgentStatusRequest struct {
	Results []CheckResultItem `json:"results" binding:"required,dive"`
}

type AgentStatusResponse struct {
	Received int `json:"received"`
}

type AgentConfigResponse struct {
	Services     []ServiceConfig `json:"services"`
	GlobalConfig GlobalConfig    `json:"global_config"`
}

type ServiceConfig struct {
	ID               uuid.UUID `json:"id"`
	Type             string    `json:"type"`
	URL              string    `json:"url"`
	CheckInterval    int       `json:"check_interval"`
	Timeout          int       `json:"timeout"`
	Retries          int       `json:"retries"`
	FailureThreshold int       `json:"failure_threshold"`
}

type GlobalConfig struct {
	DefaultInterval int `json:"default_interval"`
	DefaultTimeout  int `json:"default_timeout"`
}

type AgentHeartbeatRequest struct {
	Metadata map[string]interface{} `json:"metadata"`
}

// ─── Admin DTOs ─────────────────────────────────────────────────────

type CreateAgentGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

type AgentGroupResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"` // Only returned on creation
	CreatedAt time.Time `json:"created_at"`
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateServiceRequest struct {
	Name             string     `json:"name" binding:"required"`
	Type             string     `json:"type" binding:"required,oneof=http tcp"`
	URL              string     `json:"url" binding:"required"` // Custom validation in handler
	AgentID          *uuid.UUID `json:"agent_id"`
	CheckInterval    int        `json:"check_interval"`
	Timeout          int        `json:"timeout"`
	Retries          int        `json:"retries"`
	FailureThreshold int        `json:"failure_threshold"`
	IsPublic         bool       `json:"is_public"`
	GroupName        string     `json:"group_name"`
}

type UpdateServiceRequest struct {
	Name             *string `json:"name"`
	Type             *string `json:"type" binding:"omitempty,oneof=http tcp"`
	URL              *string `json:"url"`
	CheckInterval    *int    `json:"check_interval"`
	Timeout          *int    `json:"timeout"`
	Retries          *int    `json:"retries"`
	FailureThreshold *int    `json:"failure_threshold"`
	IsPublic         *bool   `json:"is_public"`
	GroupName        *string `json:"group_name"`
}

type CreateIncidentRequest struct {
	ServiceID   *uuid.UUID `json:"service_id"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Severity    string     `json:"severity" binding:"required,oneof=minor major critical"`
	Status      string     `json:"status" binding:"required,oneof=investigating identified monitoring resolved"`
}

type UpdateIncidentRequest struct {
	Status   string `json:"status" binding:"required,oneof=investigating identified monitoring resolved"`
	Message  string `json:"message" binding:"required"`
}

type CreateMaintenanceRequest struct {
	Title          string      `json:"title" binding:"required"`
	Description    string      `json:"description"`
	ServiceIDs     []uuid.UUID `json:"service_ids"`
	ScheduledStart time.Time   `json:"scheduled_start" binding:"required"`
	ScheduledEnd   time.Time   `json:"scheduled_end" binding:"required"`
}

// ─── Public DTOs ────────────────────────────────────────────────────

type PublicServiceStatus struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Status    string             `json:"status"`
	GroupName string             `json:"group_name"`
	UptimePct float64            `json:"uptime_pct"`
	History   []UptimeDayStatus  `json:"history,omitempty"`
}

type UptimeDayStatus struct {
	Date      string  `json:"date"`
	UptimePct float64 `json:"uptime_pct"`
	Status    string  `json:"status"` // up, down, degraded, partial
}

// ─── Generic Responses ──────────────────────────────────────────────

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
