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
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	utils.InitLogger()
	defer utils.Sync()

	utils.Info("Starting RunSight API server", zap.String("version", "1.0.0"))

	db, err := database.Connect()
	if err != nil {
		utils.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := database.Migrate(db); err != nil {
		utils.Fatal("Failed to run migrations", zap.Error(err))
	}

	utils.Info("Database connected and migrations completed")

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Recovery())

	r.Use(middleware.RequestIDMiddleware())

	r.Use(middleware.LoggingMiddleware())

	// Set max request size to 10MB
	r.Use(middleware.MaxRequestSize(10 * 1024 * 1024))

	// Global rate limit allows 100 requests per burst and 200 per window
	r.Use(middleware.RateLimitMiddleware(100, 200))

	authHandler := handlers.NewAuthHandler(db)
	mobileHandler := handlers.NewMobileHandler(db)
	iotHandler := handlers.NewIoTHandler(db)
	monitoringHandler := handlers.NewMonitoringHandler(db)

	api := r.Group("/api/v1")
	{
		api.GET("/health", monitoringHandler.Health)
		api.GET("/health/detailed", monitoringHandler.HealthDetailed)
		api.GET("/ready", monitoringHandler.Ready)
		api.GET("/live", monitoringHandler.Live)
		api.GET("/metrics", monitoringHandler.Metrics)

		auth := api.Group("/auth")
		auth.Use(middleware.StrictRateLimitMiddleware(20))
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protectedAuth := api.Group("/auth")
		protectedAuth.Use(middleware.AuthMiddleware())
		{
			protectedAuth.GET("/profile", authHandler.GetProfile)
			protectedAuth.PUT("/profile", authHandler.UpdateProfile)
		}

		mobile := api.Group("/mobile")
		mobile.Use(middleware.AuthMiddleware())
		{
			pairing := mobile.Group("/pairing")
			pairing.Use(middleware.StrictRateLimitMiddleware(20))
			{
				pairing.POST("/request", mobileHandler.RequestPairingCode)
				pairing.GET("/:session_id/status", mobileHandler.CheckPairingStatus)
			}

			mobile.GET("/devices", mobileHandler.GetDevices)
			mobile.DELETE("/devices/:device_id", mobileHandler.RemoveDevice)

			mobile.GET("/runs", mobileHandler.ListRuns)
			mobile.GET("/runs/:run_id", mobileHandler.GetRun)
			mobile.PATCH("/runs/:run_id", mobileHandler.UpdateRunNotes)
			mobile.GET("/stats", mobileHandler.GetStats)
		}

		iot := api.Group("/iot")
		{
			iot.POST("/pairing/verify", middleware.StrictRateLimitMiddleware(5), iotHandler.VerifyPairingCode)

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

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

	// Graceful shutdown captures SIGINT and SIGTERM and allows 30s for cleanup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.Error("Server forced to shutdown", zap.Error(err))
	}

	utils.Info("Server shutdown complete")
}