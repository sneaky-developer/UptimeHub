package heartbeat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Heartbeat manages periodic heartbeats to the Master server
type Heartbeat struct {
	client    *http.Client
	masterURL string
	token     string
	interval  time.Duration
	stopCh    chan struct{}
}

// NewHeartbeat creates a new Heartbeat manager
func NewHeartbeat(masterURL, token string, interval time.Duration) *Heartbeat {
	return &Heartbeat{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		masterURL: masterURL,
		token:     token,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// Start begins sending periodic heartbeats
func (h *Heartbeat) Start() {
	go func() {
		// Send initial heartbeat
		h.send()

		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.send()
			case <-h.stopCh:
				log.Println("💓 Heartbeat stopped")
				return
			}
		}
	}()

	log.Printf("💓 Heartbeat started (every %s)", h.interval)
}

// Stop stops the heartbeat loop
func (h *Heartbeat) Stop() {
	close(h.stopCh)
}

func (h *Heartbeat) send() {
	payload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling heartbeat: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, h.masterURL+"/api/agent/heartbeat", bytes.NewReader(data))
	if err != nil {
		log.Printf("Error creating heartbeat request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("⚠️  Heartbeat failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️  Heartbeat rejected (status %d): %s", resp.StatusCode, string(body))
		return
	}

	log.Println("💓 Heartbeat sent")
}

// SendWithRetry sends a heartbeat with retry logic for resilience
func (h *Heartbeat) SendWithRetry(maxRetries int) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		payload := map[string]interface{}{
			"metadata": map[string]interface{}{
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"attempt":   attempt + 1,
			},
		}

		data, _ := json.Marshal(payload)
		req, err := http.NewRequest(http.MethodPost, h.masterURL+"/api/agent/heartbeat", bytes.NewReader(data))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+h.token)

		resp, err := h.client.Do(req)
		if err != nil {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("⚠️  Heartbeat attempt %d failed, retrying in %s: %v", attempt+1, backoff, err)
			time.Sleep(backoff)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		return fmt.Errorf("heartbeat rejected with status %d", resp.StatusCode)
	}

	return fmt.Errorf("heartbeat failed after %d retries", maxRetries)
}
