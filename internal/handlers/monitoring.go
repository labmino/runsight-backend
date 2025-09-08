package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/labmino/runsight-backend/internal/utils"
)

type MonitoringHandler struct {
	db        *gorm.DB
	startTime time.Time
}

func NewMonitoringHandler(db *gorm.DB) *MonitoringHandler {
	return &MonitoringHandler{
		db:        db,
		startTime: time.Now(),
	}
}

// Health endpoint for basic health checks
func (h *MonitoringHandler) Health(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, "RunSight API is running", gin.H{
		"status": "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Detailed health check endpoint
func (h *MonitoringHandler) HealthDetailed(c *gin.Context) {
	// Check database connectivity
	sqlDB, err := h.db.DB()
	dbStatus := "ok"
	if err != nil {
		dbStatus = "error: " + err.Error()
	} else if err = sqlDB.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	// Get database stats
	var dbStats map[string]interface{}
	if sqlDB != nil {
		stats := sqlDB.Stats()
		dbStats = map[string]interface{}{
			"open_connections":     stats.OpenConnections,
			"in_use":              stats.InUse,
			"idle":                stats.Idle,
			"wait_count":          stats.WaitCount,
			"wait_duration":       stats.WaitDuration.String(),
			"max_idle_closed":     stats.MaxIdleClosed,
			"max_idle_time_closed": stats.MaxIdleTimeClosed,
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		}
	}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
		"version":   "1.0.0", // You can make this dynamic
		"database": gin.H{
			"status": dbStatus,
			"stats":  dbStats,
		},
		"memory": gin.H{
			"alloc":         memStats.Alloc,
			"total_alloc":   memStats.TotalAlloc,
			"sys":          memStats.Sys,
			"num_gc":       memStats.NumGC,
			"goroutines":   runtime.NumGoroutine(),
		},
		"system": gin.H{
			"num_cpu":      runtime.NumCPU(),
			"go_version":   runtime.Version(),
		},
	}

	statusCode := http.StatusOK
	if dbStatus != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	utils.SuccessResponse(c, statusCode, "Detailed health status", response)
}

// Readiness endpoint for Kubernetes readiness probe
func (h *MonitoringHandler) Ready(c *gin.Context) {
	// Check if the application is ready to serve requests
	// This includes database connectivity and any other critical dependencies
	
	sqlDB, err := h.db.DB()
	if err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Database connection failed", err.Error())
		return
	}

	if err = sqlDB.Ping(); err != nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "Database ping failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Service is ready", gin.H{
		"status": "ready",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Liveness endpoint for Kubernetes liveness probe
func (h *MonitoringHandler) Live(c *gin.Context) {
	// Check if the application is alive (basic process health)
	// This should be a lightweight check
	
	utils.SuccessResponse(c, http.StatusOK, "Service is alive", gin.H{
		"status": "alive",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime": time.Since(h.startTime).String(),
	})
}

// Metrics endpoint for basic application metrics
func (h *MonitoringHandler) Metrics(c *gin.Context) {
	// Get database record counts
	var userCount, deviceCount, runCount int64
	
	h.db.Model(&struct{ ID string `gorm:"primaryKey"`}{}).Table("users").Count(&userCount)
	h.db.Model(&struct{ ID string `gorm:"primaryKey"`}{}).Table("devices").Count(&deviceCount)
	h.db.Model(&struct{ ID string `gorm:"primaryKey"`}{}).Table("runs").Count(&runCount)

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get database connection stats
	var dbConnStats map[string]interface{}
	if sqlDB, err := h.db.DB(); err == nil {
		stats := sqlDB.Stats()
		dbConnStats = map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":          stats.InUse,
			"idle":            stats.Idle,
			"wait_count":      stats.WaitCount,
		}
	}

	metrics := gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
		"application": gin.H{
			"users_total":   userCount,
			"devices_total": deviceCount,
			"runs_total":    runCount,
		},
		"system": gin.H{
			"memory_alloc_bytes":    memStats.Alloc,
			"memory_total_alloc_bytes": memStats.TotalAlloc,
			"memory_sys_bytes":      memStats.Sys,
			"gc_cycles_total":       memStats.NumGC,
			"goroutines_total":      runtime.NumGoroutine(),
			"cpu_cores":            runtime.NumCPU(),
		},
		"database": dbConnStats,
	}

	utils.SuccessResponse(c, http.StatusOK, "Application metrics", metrics)
}