package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sneaky-developer/UptimeHub/master/internal/dto"
	"github.com/sneaky-developer/UptimeHub/master/internal/models"
	"github.com/sneaky-developer/UptimeHub/master/internal/services"
)

// AgentHandler handles agent-related API endpoints
type AgentHandler struct {
	db      *gorm.DB
	monitor *services.MonitorService
}

// NewAgentHandler creates a new AgentHandler
func NewAgentHandler(db *gorm.DB, monitor *services.MonitorService) *AgentHandler {
	return &AgentHandler{db: db, monitor: monitor}
}

// Register handles POST /api/agent/register
func (h *AgentHandler) Register(c *gin.Context) {
	var req dto.AgentRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Missing or invalid enrollment token"})
		return
	}
	enrollmentToken := authHeader[7:]

	var group models.AgentGroup
	if err := h.db.Where("token = ?", enrollmentToken).First(&group).Error; err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid enrollment token"})
		return
	}

	// Generate a random session token for the agent
	token, err := generateToken(64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	var agent models.Agent
	if err := h.db.Where("name = ? AND agent_group_id = ?", req.Name, group.ID).First(&agent).Error; err != nil {
		// Does not exist, create new
		agent = models.Agent{
			Name:         req.Name,
			AgentGroupID: &group.ID,
			Token:        token,
			Status:       "active",
			Metadata:     models.JSON(req.Metadata),
		}
		if err := h.db.Create(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to register new agent"})
			return
		}
	} else {
		// Update existing agent
		agent.Token = token
		agent.Metadata = models.JSON(req.Metadata)
		agent.Status = "active"
		if err := h.db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to update existing agent"})
			return
		}
	}

	c.JSON(http.StatusCreated, dto.AgentRegisterResponse{
		AgentID: agent.ID,
		Token:   token,
	})
}

// Status handles POST /api/agent/status
func (h *AgentHandler) Status(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	var req dto.AgentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	// Convert DTOs to models
	checkResults := make([]models.CheckResult, 0, len(req.Results))
	for _, r := range req.Results {
		checkResults = append(checkResults, models.CheckResult{
			ServiceID:    r.ServiceID,
			StatusCode:   r.StatusCode,
			ResponseTime: r.ResponseTime,
			IsUp:         r.IsUp,
			ErrorMessage: r.ErrorMessage,
			CheckedAt:    r.CheckedAt,
		})
	}

	received, err := h.monitor.ProcessCheckResults(agentID, checkResults)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to process results"})
		return
	}

	c.JSON(http.StatusOK, dto.AgentStatusResponse{Received: received})
}

// Config handles GET /api/agent/config
func (h *AgentHandler) Config(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	var svcs []models.Service
	h.db.Where("agent_id = ?", agentID).Find(&svcs)

	serviceConfigs := make([]dto.ServiceConfig, 0, len(svcs))
	for _, svc := range svcs {
		serviceConfigs = append(serviceConfigs, dto.ServiceConfig{
			ID:               svc.ID,
			Type:             svc.Type,
			URL:              svc.URL,
			CheckInterval:    svc.CheckInterval,
			Timeout:          svc.Timeout,
			Retries:          svc.Retries,
			FailureThreshold: svc.FailureThreshold,
		})
	}

	c.JSON(http.StatusOK, dto.AgentConfigResponse{
		Services: serviceConfigs,
		GlobalConfig: dto.GlobalConfig{
			DefaultInterval: 30,
			DefaultTimeout:  10,
		},
	})
}

// Discovery handles POST /api/agent/discovery — agents report K8s services
// found via label-based discovery. The master upserts them as services owned
// by the reporting agent so they flow back through GET /api/agent/config.
func (h *AgentHandler) Discovery(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	var req dto.AgentDiscoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	created, updated := 0, 0
	keys := make([]string, 0, len(req.Services))

	for _, item := range req.Services {
		keys = append(keys, item.Key)

		var svc models.Service
		err := h.db.Where("agent_id = ? AND discovery_key = ?", agentID, item.Key).First(&svc).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			name := item.Name
			if item.Namespace != "" {
				name = item.Namespace + "/" + item.Name
			}
			svc = models.Service{
				AgentID:      &agentID,
				Name:         name,
				Type:         "http",
				URL:          item.URL,
				Source:       "discovered",
				DiscoveryKey: item.Key,
				GroupName:    item.Namespace,
				IsPublic:     false, // admin opts in to exposing on the status page
			}
			if err := h.db.Create(&svc).Error; err != nil {
				log.Printf("Error creating discovered service %s: %v", item.Key, err)
				continue
			}
			created++
		case err == nil:
			if svc.URL != item.URL {
				h.db.Model(&svc).Update("url", item.URL)
				updated++
			}
		default:
			log.Printf("Error looking up discovered service %s: %v", item.Key, err)
		}
	}

	// Prune discovered services that disappeared from the cluster — only when
	// the agent confirms the list is complete, so a partial discovery (e.g.
	// missing ingress RBAC) never wipes valid services.
	pruned := 0
	if req.Complete {
		q := h.db.Where("agent_id = ? AND source = 'discovered'", agentID)
		if len(keys) > 0 {
			q = q.Where("discovery_key NOT IN ?", keys)
		}
		result := q.Delete(&models.Service{})
		pruned = int(result.RowsAffected)
		if pruned > 0 {
			log.Printf("🧹 Pruned %d discovered services no longer present for agent %s", pruned, agentID)
		}
	}

	c.JSON(http.StatusOK, dto.AgentDiscoveryResponse{Created: created, Updated: updated, Pruned: pruned})
}

// Heartbeat handles POST /api/agent/heartbeat
func (h *AgentHandler) Heartbeat(c *gin.Context) {
	agentID := c.MustGet("agent_id").(uuid.UUID)

	var req dto.AgentHeartbeatRequest
	c.ShouldBindJSON(&req)

	now := time.Now()
	updates := map[string]interface{}{
		"last_heartbeat": now,
		"status":         "active",
	}
	if req.Metadata != nil {
		updates["metadata"] = models.JSON(req.Metadata)
	}

	h.db.Model(&models.Agent{}).Where("id = ?", agentID).Updates(updates)

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "heartbeat received"})
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
