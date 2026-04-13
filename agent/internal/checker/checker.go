package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// CheckResult holds the result of a single health check
type CheckResult struct {
	ServiceID    string
	URL          string
	StatusCode   int
	ResponseTime int // milliseconds
	IsUp         bool
	ErrorMessage string
	CheckedAt    time.Time
}

// Target represents an endpoint to be health-checked
type Target struct {
	ServiceID        string
	Type             string
	URL              string
	Timeout          time.Duration
	Retries          int
	FailureThreshold int
}

// Checker performs HTTP health checks
type Checker struct {
	client *http.Client
}

// NewChecker creates a new Checker
func NewChecker(defaultTimeout time.Duration) *Checker {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &Checker{
		client: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
			// Don't follow redirects — we want the actual status
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

// CheckAll runs health checks for all targets concurrently
func (c *Checker) CheckAll(ctx context.Context, targets []Target) []CheckResult {
	var wg sync.WaitGroup
	results := make([]CheckResult, len(targets))

	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t Target) {
			defer wg.Done()
			results[idx] = c.checkWithRetries(ctx, t)
		}(i, target)
	}

	wg.Wait()
	return results
}

// checkWithRetries performs a health check with retries
func (c *Checker) checkWithRetries(ctx context.Context, target Target) CheckResult {
	retries := target.Retries
	if retries < 1 {
		retries = 1
	}

	var lastResult CheckResult

	for attempt := 0; attempt < retries; attempt++ {
		lastResult = c.check(ctx, target)

		if lastResult.IsUp {
			return lastResult
		}

		if attempt < retries-1 {
			// Wait before retry (exponential backoff: 1s, 2s, 4s...)
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				lastResult.ErrorMessage = "context cancelled"
				return lastResult
			case <-time.After(backoff):
			}
		}
	}

	log.Printf("❌ Health check failed after %d retries: %s — %s", retries, target.URL, lastResult.ErrorMessage)
	return lastResult
}

// check performs a single HTTP health check
func (c *Checker) check(ctx context.Context, target Target) CheckResult {
	if target.Type == "tcp" {
		return c.checkTCP(ctx, target)
	}
	return c.checkHTTP(ctx, target)
}

// checkTCP performs a raw TCP port check
func (c *Checker) checkTCP(ctx context.Context, target Target) CheckResult {
	result := CheckResult{
		ServiceID: target.ServiceID,
		URL:       target.URL,
		CheckedAt: time.Now().UTC(),
	}

	timeout := target.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	start := time.Now()
	
	// Create a dialer with timeout
	dialer := net.Dialer{Timeout: timeout}
	
	// DialContext allows the connection attempt to be cancelled via context
	conn, err := dialer.DialContext(ctx, "tcp", target.URL)
	elapsed := time.Since(start)
	
	result.ResponseTime = int(elapsed.Milliseconds())

	if err != nil {
		result.IsUp = false
		result.ErrorMessage = fmt.Sprintf("TCP connection failed: %v", err)
		return result
	}
	
	defer conn.Close()
	result.IsUp = true
	result.StatusCode = 200 // Mock status code for TCP success

	return result
}

// checkHTTP performs a single HTTP health check
func (c *Checker) checkHTTP(ctx context.Context, target Target) CheckResult {
	result := CheckResult{
		ServiceID: target.ServiceID,
		URL:       target.URL,
		CheckedAt: time.Now().UTC(),
	}

	// Create request with timeout
	client := c.client
	if target.Timeout > 0 {
		client = &http.Client{
			Timeout:       target.Timeout,
			Transport:     c.client.Transport,
			CheckRedirect: c.client.CheckRedirect,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	if err != nil {
		result.IsUp = false
		result.ErrorMessage = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	req.Header.Set("User-Agent", "UptimeHub-Agent/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	result.ResponseTime = int(elapsed.Milliseconds())

	if err != nil {
		result.IsUp = false
		result.ErrorMessage = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.IsUp = resp.StatusCode >= 200 && resp.StatusCode < 400

	if !result.IsUp {
		result.ErrorMessage = fmt.Sprintf("unhealthy status code: %d", resp.StatusCode)
	}

	return result
}
