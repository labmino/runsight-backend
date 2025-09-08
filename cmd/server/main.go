package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/labmino/runsight-backend/internal/database"
	"github.com/labmino/runsight-backend/internal/handlers"
	"github.com/labmino/runsight-backend/internal/middleware"
	"github.com/labmino/runsight-backend/internal/utils"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize structured logging
	utils.InitLogger()
	defer utils.Sync()

	utils.Info("Starting RunSight API server", zap.String("version", "1.0.0"))

	// Connect to database with retries
	db, err := database.Connect()
	if err != nil {
		utils.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		utils.Fatal("Failed to run migrations", zap.Error(err))
	}

	utils.Info("Database connected and migrations completed")

	// Initialize Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Recovery middleware
	r.Use(gin.Recovery())

	// Request ID middleware (must be first)
	r.Use(middleware.RequestIDMiddleware())

	// Structured logging middleware
	r.Use(middleware.LoggingMiddleware())

	// Request size limit (still needed for mobile apps)
	r.Use(middleware.MaxRequestSize(10 * 1024 * 1024)) // 10MB max

	// Note: CORS not needed for mobile apps, only for web browsers

	// Rate limiting
	r.Use(middleware.RateLimitMiddleware(100, 200)) // 100 req/sec, burst 200

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	mobileHandler := handlers.NewMobileHandler(db)
	iotHandler := handlers.NewIoTHandler(db)
	monitoringHandler := handlers.NewMonitoringHandler(db)

	// API routes
	api := r.Group("/api/v1")
	{
		// Health and monitoring endpoints
		api.GET("/health", monitoringHandler.Health)
		api.GET("/health/detailed", monitoringHandler.HealthDetailed)
		api.GET("/ready", monitoringHandler.Ready)
		api.GET("/live", monitoringHandler.Live)
		api.GET("/metrics", monitoringHandler.Metrics)

		// Authentication endpoints with strict rate limiting
		auth := api.Group("/auth")
		auth.Use(middleware.StrictRateLimitMiddleware(20)) // 20 requests per minute
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected auth endpoints
		protectedAuth := api.Group("/auth")
		protectedAuth.Use(middleware.AuthMiddleware())
		{
			protectedAuth.GET("/profile", authHandler.GetProfile)
			protectedAuth.PUT("/profile", authHandler.UpdateProfile)
		}

		// Mobile app endpoints
		mobile := api.Group("/mobile")
		mobile.Use(middleware.AuthMiddleware())
		{
			// Pairing endpoints with stricter rate limiting
			pairing := mobile.Group("/pairing")
			pairing.Use(middleware.StrictRateLimitMiddleware(20)) // 20 requests per minute for pairing (allows polling every 5s)
			{
				pairing.POST("/request", mobileHandler.RequestPairingCode)
				pairing.GET("/:session_id/status", mobileHandler.CheckPairingStatus)
			}

			// Device management
			mobile.GET("/devices", mobileHandler.GetDevices)
			mobile.DELETE("/devices/:device_id", mobileHandler.RemoveDevice)

			// Run management
			mobile.GET("/runs", mobileHandler.ListRuns)
			mobile.GET("/runs/:run_id", mobileHandler.GetRun)
			mobile.PATCH("/runs/:run_id", mobileHandler.UpdateRunNotes)
			mobile.GET("/stats", mobileHandler.GetStats)
		}

		// IoT device endpoints
		iot := api.Group("/iot")
		{
			// Pairing verification with strict rate limiting
			iot.POST("/pairing/verify", middleware.StrictRateLimitMiddleware(5), iotHandler.VerifyPairingCode)

			// Protected IoT endpoints
			iotProtected := iot.Group("")
			iotProtected.Use(middleware.DeviceAuthMiddleware(db))
			{
				iotProtected.POST("/runs/upload", iotHandler.UploadRun)
				iotProtected.POST("/runs/batch", iotHandler.BatchUploadRuns)
				iotProtected.POST("/devices/status", iotHandler.UpdateDeviceStatus)
				iotProtected.GET("/devices/config", iotHandler.GetDeviceConfig)
			}
		}
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Parse timeouts from environment
	readTimeout := 30 * time.Second
	writeTimeout := 30 * time.Second
	idleTimeout := 120 * time.Second

	if timeoutStr := os.Getenv("READ_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			readTimeout = timeout
		}
	}

	if timeoutStr := os.Getenv("WRITE_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			writeTimeout = timeout
		}
	}

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Start server in a goroutine
	go func() {
		utils.Info("Starting HTTP server", 
			zap.String("port", port),
			zap.Duration("read_timeout", readTimeout),
			zap.Duration("write_timeout", writeTimeout),
			zap.Duration("idle_timeout", idleTimeout),
		)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.Error("Server forced to shutdown", zap.Error(err))
	}

	utils.Info("Server shutdown complete")
}