package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sneaky-developer/UptimeHub/agent/internal/checker"
	"github.com/sneaky-developer/UptimeHub/agent/internal/config"
	"github.com/sneaky-developer/UptimeHub/agent/internal/discovery"
	"github.com/sneaky-developer/UptimeHub/agent/internal/heartbeat"
	"github.com/sneaky-developer/UptimeHub/agent/internal/reporter"
)

func main() {
	log.Println("🚀 UptimeHub Agent starting...")

	// Load configuration
	cfg := config.Load()

	// Initialize reporter (Master communication)
	rep := reporter.NewReporter(cfg.MasterURL, cfg.AgentToken)

	// Register with Master if no token provided
	if cfg.AgentToken == "" {
		log.Println("📝 No agent token found, registering with Master...")

		result, err := registerWithRetry(rep, cfg, 5)
		if err != nil {
			log.Fatalf("❌ Failed to register with Master: %v", err)
		}

		cfg.AgentToken = result.Token
		rep.SetToken(result.Token)
		log.Printf("✅ Registered as agent %s", result.AgentID)
	}

	// Initialize Kubernetes discovery (optional — may fail outside K8s)
	var disc *discovery.Discovery
	disc, err := discovery.NewDiscovery(cfg.InCluster, cfg.KubeNamespace)
	if err != nil {
		log.Printf("⚠️  Kubernetes discovery unavailable: %v", err)
		log.Println("   Agent will only monitor services configured in Master")
	}

	// Initialize health checker
	chk := checker.NewChecker(cfg.DefaultTimeout)

	// Start heartbeat
	hb := heartbeat.NewHeartbeat(cfg.MasterURL, cfg.AgentToken, cfg.HeartbeatInterval)
	hb.Start()

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Main monitoring loop
	go runMonitoringLoop(ctx, cfg, disc, chk, rep)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down agent...")
	cancel()
	hb.Stop()
	time.Sleep(time.Second) // Allow goroutines to finish
	log.Println("✅ Agent stopped")
}

func runMonitoringLoop(ctx context.Context, cfg *config.Config, disc *discovery.Discovery, chk *checker.Checker, rep *reporter.Reporter) {
	// Fetch initial config from Master
	masterConfig, err := rep.FetchConfig()
	if err != nil {
		log.Printf("⚠️  Failed to fetch initial config from Master: %v", err)
	}

	configSyncTicker := time.NewTicker(cfg.ConfigSyncInterval)
	defer configSyncTicker.Stop()

	checkTicker := time.NewTicker(cfg.DefaultInterval)
	defer checkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-configSyncTicker.C:
			// Resync config from Master
			newConfig, err := rep.FetchConfig()
			if err != nil {
				log.Printf("⚠️  Config sync failed: %v", err)
			} else {
				masterConfig = newConfig
				log.Printf("🔄 Config synced: %d services", len(masterConfig.Services))
			}

			// Re-discover K8s services
			if disc != nil {
				_, err := disc.Discover(ctx)
				if err != nil {
					log.Printf("⚠️  K8s discovery failed: %v", err)
				}
			}

		case <-checkTicker.C:
			// Build targets from Master config
			var targets []checker.Target

			if masterConfig != nil {
				for _, svc := range masterConfig.Services {
					targets = append(targets, checker.Target{
						ServiceID: svc.ID,
						URL:       svc.URL,
						Timeout:   time.Duration(svc.Timeout) * time.Second,
						Retries:   svc.Retries,
					})
				}
			}

			if len(targets) == 0 {
				continue
			}

			// Run health checks
			results := chk.CheckAll(ctx, targets)

			// Convert to reporter format and send
			var payloads []reporter.CheckResultPayload
			for _, r := range results {
				payloads = append(payloads, reporter.CheckResultPayload{
					ServiceID:    r.ServiceID,
					URL:          r.URL,
					StatusCode:   r.StatusCode,
					ResponseTime: r.ResponseTime,
					IsUp:         r.IsUp,
					ErrorMessage: r.ErrorMessage,
					CheckedAt:    r.CheckedAt,
				})
			}

			if err := rep.ReportResults(payloads); err != nil {
				log.Printf("⚠️  Failed to report results: %v", err)
			}
		}
	}
}

func registerWithRetry(rep *reporter.Reporter, cfg *config.Config, maxRetries int) (*reporter.RegisterResponse, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := rep.Register(cfg.AgentName, cfg.ClusterName, map[string]interface{}{
			"version": "1.0.0",
		})
		if err == nil {
			return result, nil
		}

		lastErr = err
		backoff := time.Duration(1<<uint(attempt)) * time.Second
		log.Printf("⚠️  Registration attempt %d failed, retrying in %s: %v", attempt+1, backoff, err)
		time.Sleep(backoff)
	}

	return nil, lastErr
}
