package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/sneaky-developer/UptimeHub/master/internal/config"
	"github.com/sneaky-developer/UptimeHub/master/internal/database"
	"github.com/sneaky-developer/UptimeHub/master/internal/handlers"
	"github.com/sneaky-developer/UptimeHub/master/internal/middleware"
	"github.com/sneaky-developer/UptimeHub/master/internal/notifier"
	"github.com/sneaky-developer/UptimeHub/master/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Seed default admin user
	services.SeedDefaultAdmin(db)

	// Initialize services
	notifierSvc := notifier.NewService(db)
	monitorSvc := services.NewMonitorService(db, notifierSvc)
	aggregatorSvc := services.NewAggregatorService(db)

	// Start background workers
	aggregatorSvc.Start()

	// Initialize handlers
	agentHandler := handlers.NewAgentHandler(db, monitorSvc)
	adminHandler := handlers.NewAdminHandler(db, cfg.JWTSecret)
	publicHandler := handlers.NewPublicHandler(db)

	// Setup router
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "uptimehub-master"})
	})

	// ─── Public API (no auth) ───────────────────────────────────────
	public := r.Group("/api")
	{
		public.GET("/status", publicHandler.Status)
		public.GET("/status/:id/history", publicHandler.ServiceHistory)
		public.GET("/incidents", publicHandler.Incidents)
		public.GET("/maintenance", publicHandler.Maintenance)
	}

	// ─── Agent API (agent token auth) ───────────────────────────────
	// Registration endpoint (no auth — agents register to get a token)
	r.POST("/api/agent/register", agentHandler.Register)

	agent := r.Group("/api/agent")
	agent.Use(middleware.AgentAuth(db))
	{
		agent.POST("/status", agentHandler.Status)
		agent.GET("/config", agentHandler.Config)
		agent.POST("/heartbeat", agentHandler.Heartbeat)
	}

	// ─── Admin API (JWT auth) ───────────────────────────────────────
	r.POST("/api/admin/login", adminHandler.Login)

	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Agents
		admin.GET("/agents", adminHandler.ListAgents)

		// Services
		admin.GET("/services", adminHandler.ListServices)
		admin.POST("/services", adminHandler.CreateService)
		admin.PUT("/services/:id", adminHandler.UpdateService)
		admin.DELETE("/services/:id", adminHandler.DeleteService)

		// Incidents
		admin.GET("/incidents", adminHandler.ListIncidents)
		admin.POST("/incidents", adminHandler.CreateIncident)
		admin.PUT("/incidents/:id", adminHandler.UpdateIncident)

		// Maintenance
		admin.GET("/maintenance", adminHandler.ListMaintenance)
		admin.POST("/maintenance", adminHandler.CreateMaintenance)

		// Config
		admin.PUT("/config", adminHandler.UpdateConfig)

		// ─── Alerts ─────────────────────────────────────────────────────
		alertHandler := handlers.NewAlertHandler(db, notifierSvc)
		alerts := admin.Group("/alerts")
		{
			alerts.GET("/channels", alertHandler.ListChannels)
			alerts.POST("/channels", alertHandler.CreateChannel)
			alerts.PUT("/channels/:id", alertHandler.UpdateChannel)
			alerts.DELETE("/channels/:id", alertHandler.DeleteChannel)
			alerts.POST("/channels/:id/test", alertHandler.TestChannel)
		}
	}

	// ─── Graceful shutdown ──────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 UptimeHub Master starting on :%s", cfg.AppPort)
		if err := r.Run(":" + cfg.AppPort); err != nil {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("🛑 Shutting down server...")

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Println("✅ Server stopped gracefully")
}
