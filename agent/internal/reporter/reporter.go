package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Reporter sends health check results to the Master server
type Reporter struct {
	client    *http.Client
	masterURL string
	token     string
}

// NewReporter creates a new Reporter
func NewReporter(masterURL, token string) *Reporter {
	return &Reporter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		masterURL: masterURL,
		token:     token,
	}
}

// SetToken updates the authentication token
func (r *Reporter) SetToken(token string) {
	r.token = token
}

// CheckResultPayload matches the Master's expected format
type CheckResultPayload struct {
	ServiceID    string    `json:"service_id"`
	URL          string    `json:"url"`
	StatusCode   int       `json:"status_code"`
	ResponseTime int       `json:"response_time_ms"`
	IsUp         bool      `json:"is_up"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}

// StatusRequest is the batch request body
type StatusRequest struct {
	Results []CheckResultPayload `json:"results"`
}

// ReportResults sends a batch of check results to the Master
func (r *Reporter) ReportResults(results []CheckResultPayload) error {
	body := StatusRequest{Results: results}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, r.masterURL+"/api/agent/status", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("master returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("📤 Reported %d check results to master", len(results))
	return nil
}

// RegisterRequest is the body for agent registration
type RegisterRequest struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RegisterResponse holds the registration response
type RegisterResponse struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

// Register registers the agent with the Master and returns the session token
func (r *Reporter) Register(name, enrollmentToken string, metadata map[string]interface{}) (*RegisterResponse, error) {
	body := RegisterRequest{
		Name:     name,
		Metadata: metadata,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, r.masterURL+"/api/agent/register", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+enrollmentToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registration failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	log.Printf("✅ Agent registered: %s (ID: %s)", name, result.AgentID)
	return &result, nil
}

// ServiceConfig from Master
type ServiceConfig struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	URL              string `json:"url"`
	CheckInterval    int    `json:"check_interval"`
	Timeout          int    `json:"timeout"`
	Retries          int    `json:"retries"`
	FailureThreshold int    `json:"failure_threshold"`
}

// ConfigResponse from Master
type ConfigResponse struct {
	Services     []ServiceConfig `json:"services"`
	GlobalConfig struct {
		DefaultInterval int `json:"default_interval"`
		DefaultTimeout  int `json:"default_timeout"`
	} `json:"global_config"`
}

// FetchConfig retrieves the current config from the Master
func (r *Reporter) FetchConfig() (*ConfigResponse, error) {
	req, err := http.NewRequest(http.MethodGet, r.masterURL+"/api/agent/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("master returned status %d", resp.StatusCode)
	}

	var config ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}
