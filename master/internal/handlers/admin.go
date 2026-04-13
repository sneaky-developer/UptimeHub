package handlers

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/dto"
	"github.com/sneaky-developer/UptimeHub/master/internal/middleware"
	"github.com/sneaky-developer/UptimeHub/master/internal/models"
)

// AdminHandler handles admin-related API endpoints
type AdminHandler struct {
	db        *gorm.DB
	jwtSecret string
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(db *gorm.DB, jwtSecret string) *AdminHandler {
	return &AdminHandler{db: db, jwtSecret: jwtSecret}
}

// Login handles POST /api/admin/login
func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	var user models.AdminUser
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	token, err := middleware.GenerateJWT(&user, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, dto.AdminLoginResponse{
		Token: token,
		Name:  user.Name,
		Email: user.Email,
	})
}

// ─── Agent Groups ───────────────────────────────────────────────────

// ListAgentGroups handles GET /api/admin/agent-groups
func (h *AdminHandler) ListAgentGroups(c *gin.Context) {
	var groups []models.AgentGroup
	h.db.Order("created_at DESC").Find(&groups)

	responses := make([]dto.AgentGroupResponse, 0, len(groups))
	for _, g := range groups {
		responses = append(responses, dto.AgentGroupResponse{
			ID:        g.ID,
			Name:      g.Name,
			CreatedAt: g.CreatedAt,
			// Intentionally omitting Token for list view
		})
	}

	c.JSON(http.StatusOK, responses)
}

// CreateAgentGroup handles POST /api/admin/agent-groups
func (h *AdminHandler) CreateAgentGroup(c *gin.Context) {
	var req dto.CreateAgentGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	token, err := generateToken(64) // generateToken is defined in agent.go
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	group := models.AgentGroup{
		Name:  req.Name,
		Token: token,
	}

	if err := h.db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create agent group (name might already exist)"})
		return
	}

	c.JSON(http.StatusCreated, dto.AgentGroupResponse{
		ID:        group.ID,
		Name:      group.Name,
		Token:     group.Token, // Return token ONLY on creation
		CreatedAt: group.CreatedAt,
	})
}

// DeleteAgentGroup handles DELETE /api/admin/agent-groups/:id
func (h *AdminHandler) DeleteAgentGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid group ID"})
		return
	}

	// Delete associated agents first
	h.db.Where("agent_group_id = ?", id).Delete(&models.Agent{})
	// Delete group
	if err := h.db.Delete(&models.AgentGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to delete agent group"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Agent group deleted"})
}

// ─── Agents ─────────────────────────────────────────────────────────

// ListAgents handles GET /api/admin/agents
func (h *AdminHandler) ListAgents(c *gin.Context) {
	var agents []models.Agent
	h.db.Preload("AgentGroup").Order("created_at DESC").Find(&agents)
	c.JSON(http.StatusOK, agents)
}

// ─── Services ───────────────────────────────────────────────────────

// ListServices handles GET /api/admin/services
func (h *AdminHandler) ListServices(c *gin.Context) {
	var services []models.Service
	h.db.Preload("Agent").Order("group_name, name").Find(&services)
	c.JSON(http.StatusOK, services)
}

// CreateService handles POST /api/admin/services
func (h *AdminHandler) CreateService(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := validateServiceURL(req.Type, req.URL); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid URL/Address format for select type", Message: err.Error()})
		return
	}

	service := models.Service{
		Name:             req.Name,
		Type:             req.Type,
		URL:              req.URL,
		AgentID:          req.AgentID,
		CheckInterval:    defaultInt(req.CheckInterval, 30),
		Timeout:          defaultInt(req.Timeout, 10),
		Retries:          defaultInt(req.Retries, 3),
		FailureThreshold: defaultInt(req.FailureThreshold, 3),
		IsPublic:         req.IsPublic,
		GroupName:        req.GroupName,
		Status:           "unknown",
	}

	if err := h.db.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create service"})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// UpdateService handles PUT /api/admin/services/:id
func (h *AdminHandler) UpdateService(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid service ID"})
		return
	}

	var service models.Service
	if err := h.db.First(&service, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Service not found"})
		return
	}

	var req dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if req.URL != nil && req.Type != nil {
		if err := validateServiceURL(*req.Type, *req.URL); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid URL for selected type", Message: err.Error()})
			return
		}
	} else if req.URL != nil {
		// Verify against existing type
		if err := validateServiceURL(service.Type, *req.URL); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid URL for selected type", Message: err.Error()})
			return
		}
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.CheckInterval != nil {
		updates["check_interval"] = *req.CheckInterval
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if req.Retries != nil {
		updates["retries"] = *req.Retries
	}
	if req.FailureThreshold != nil {
		updates["failure_threshold"] = *req.FailureThreshold
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.GroupName != nil {
		updates["group_name"] = *req.GroupName
	}

	h.db.Model(&service).Updates(updates)
	h.db.First(&service, id)

	c.JSON(http.StatusOK, service)
}

// DeleteService handles DELETE /api/admin/services/:id
func (h *AdminHandler) DeleteService(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid service ID"})
		return
	}

	if err := h.db.Delete(&models.Service{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to delete service"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Service deleted"})
}

// ─── Incidents ──────────────────────────────────────────────────────

// ListIncidents handles GET /api/admin/incidents
func (h *AdminHandler) ListIncidents(c *gin.Context) {
	var incidents []models.Incident
	h.db.Preload("Updates").Preload("Service").Order("created_at DESC").Find(&incidents)
	c.JSON(http.StatusOK, incidents)
}

// CreateIncident handles POST /api/admin/incidents
func (h *AdminHandler) CreateIncident(c *gin.Context) {
	var req dto.CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	incident := models.Incident{
		ServiceID:   req.ServiceID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Severity:    req.Severity,
		StartedAt:   time.Now(),
		IsManual:    true,
	}

	if err := h.db.Create(&incident).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create incident"})
		return
	}

	// Create initial update
	update := models.IncidentUpdate{
		IncidentID: incident.ID,
		Status:     req.Status,
		Message:    req.Description,
	}
	h.db.Create(&update)

	c.JSON(http.StatusCreated, incident)
}

// UpdateIncident handles PUT /api/admin/incidents/:id
func (h *AdminHandler) UpdateIncident(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid incident ID"})
		return
	}

	var incident models.Incident
	if err := h.db.First(&incident, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Incident not found"})
		return
	}

	var req dto.UpdateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == "resolved" {
		now := time.Now()
		updates["resolved_at"] = now
	}

	h.db.Model(&incident).Updates(updates)

	// Add incident update
	update := models.IncidentUpdate{
		IncidentID: incident.ID,
		Status:     req.Status,
		Message:    req.Message,
	}
	h.db.Create(&update)

	h.db.Preload("Updates").First(&incident, id)
	c.JSON(http.StatusOK, incident)
}

// ─── Maintenance ────────────────────────────────────────────────────

// ListMaintenance handles GET /api/admin/maintenance
func (h *AdminHandler) ListMaintenance(c *gin.Context) {
	var windows []models.MaintenanceWindow
	h.db.Order("scheduled_start DESC").Find(&windows)
	c.JSON(http.StatusOK, windows)
}

// CreateMaintenance handles POST /api/admin/maintenance
func (h *AdminHandler) CreateMaintenance(c *gin.Context) {
	var req dto.CreateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	window := models.MaintenanceWindow{
		Title:          req.Title,
		Description:    req.Description,
		ServiceIDs:     models.UUIDArray(req.ServiceIDs),
		ScheduledStart: req.ScheduledStart,
		ScheduledEnd:   req.ScheduledEnd,
	}

	if err := h.db.Create(&window).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create maintenance window"})
		return
	}

	c.JSON(http.StatusCreated, window)
}

// ─── Config ─────────────────────────────────────────────────────────

// UpdateConfig handles PUT /api/admin/config
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	// For Phase 1, this updates default check parameters for all services
	var req struct {
		DefaultInterval    int `json:"default_interval"`
		DefaultTimeout     int `json:"default_timeout"`
		DefaultRetries     int `json:"default_retries"`
		FailureThreshold   int `json:"failure_threshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Global config updated",
		Data:    req,
	})
}

func defaultInt(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}

func validateServiceURL(serviceType, u string) error {
	if serviceType == "tcp" {
		host, port, err := net.SplitHostPort(u)
		if err != nil {
			return errors.New("TCP address must be in format host:port")
		}
		if port == "" {
			return errors.New("TCP address must specify a port")
		}
		if host == "" {
			return errors.New("TCP address must specify a host")
		}
		return nil
	}

	// Default to HTTP validation
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return errors.New("invalid HTTP URL format")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL scheme must be http or https")
	}
	return nil
}
