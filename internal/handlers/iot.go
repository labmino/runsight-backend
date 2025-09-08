package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/labmino/runsight-backend/internal/models"
	"github.com/labmino/runsight-backend/internal/services"
	"github.com/labmino/runsight-backend/internal/utils"
)

type IoTHandler struct {
	db             *gorm.DB
	validator      *validator.Validate
	pairingService *services.PairingService
}

func NewIoTHandler(db *gorm.DB) *IoTHandler {
	return &IoTHandler{
		db:             db,
		validator:      validator.New(),
		pairingService: services.NewPairingService(db),
	}
}

func (h *IoTHandler) VerifyPairingCode(c *gin.Context) {
	var req models.DeviceRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	device, err := h.pairingService.VerifyPairingCode(&req)
	if err != nil {
		if err.Error() == "invalid or expired pairing code" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid pairing code", err.Error())
			return
		}
		if err.Error() == "device already registered" {
			utils.ErrorResponse(c, http.StatusConflict, "Device already paired", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify pairing code", err.Error())
		return
	}

	response := gin.H{
		"device_token": device.DeviceToken,
		"user_id":      device.UserID,
		"config": models.DeviceConfig{
			UploadIntervalSeconds: 300,
			BatchSize:            10,   
			CompressionEnabled:   true,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Device paired successfully", response)
}

func (h *IoTHandler) UploadRun(c *gin.Context) {
	device, exists := c.Get("device")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Device not found in context", "")
		return
	}

	deviceInfo, ok := device.(*models.Device)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid device context", "")
		return
	}

	var req struct {
		DeviceID  string                      `json:"device_id" binding:"required"`
		SessionID string                      `json:"session_id" binding:"required"`
		RunData   models.RunCreateRequest     `json:"run_data" binding:"required"`
		AIMetrics *models.AIMetricsRequest    `json:"ai_metrics,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if req.DeviceID != deviceInfo.DeviceID {
		utils.ErrorResponse(c, http.StatusForbidden, "Device ID mismatch", "Authenticated device does not match request device_id")
		return
	}

	var existingRun models.Run
	err := h.db.Where("session_id = ?", req.SessionID).First(&existingRun).Error
	if err == nil {
		utils.SuccessResponse(c, http.StatusOK, "Run already exists", gin.H{
			"run_id": existingRun.ID,
			"status": "already_exists",
		})
		return
	}
	if err != gorm.ErrRecordNotFound {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	if err := h.validator.Struct(&req.RunData); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	run := models.Run{
		UserID:          deviceInfo.UserID,
		DeviceID:        req.DeviceID,
		SessionID:       req.SessionID,
		Title:           req.RunData.Title,
		Notes:           req.RunData.Notes,
		StartedAt:       req.RunData.StartedAt,
		EndedAt:         req.RunData.EndedAt,
		DurationSeconds: req.RunData.DurationSeconds,
		DistanceMeters:  req.RunData.DistanceMeters,
		AvgSpeedKmh:     req.RunData.AvgSpeedKmh,
		MaxSpeedKmh:     req.RunData.MaxSpeedKmh,
		CaloriesBurned:  req.RunData.CaloriesBurned,
		StepsCount:      req.RunData.StepsCount,
		StartLatitude:   req.RunData.StartLatitude,
		StartLongitude:  req.RunData.StartLongitude,
		EndLatitude:     req.RunData.EndLatitude,
		EndLongitude:    req.RunData.EndLongitude,
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to start transaction", tx.Error.Error())
		return
	}

	if err := tx.Create(&run).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create run", err.Error())
		return
	}

	if req.AIMetrics != nil {
		if err := h.validator.Struct(req.AIMetrics); err != nil {
			tx.Rollback()
			utils.ValidationErrorResponse(c, err)
			return
		}

		aiMetrics := req.AIMetrics.ToModel(run.ID)
		if err := tx.Create(aiMetrics).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save AI metrics", err.Error())
			return
		}
	}

	now := time.Now()
	deviceInfo.LastSyncAt = &now
	if err := tx.Save(deviceInfo).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update device sync time", err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Run uploaded successfully", gin.H{
		"run_id": run.ID,
		"status": "saved",
	})
}

func (h *IoTHandler) BatchUploadRuns(c *gin.Context) {
	device, exists := c.Get("device")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Device not found in context", "")
		return
	}

	deviceInfo, ok := device.(*models.Device)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid device context", "")
		return
	}

	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
		Runs     []struct {
			SessionID string                   `json:"session_id" binding:"required"`
			RunData   models.RunCreateRequest  `json:"run_data" binding:"required"`
			AIMetrics *models.AIMetricsRequest `json:"ai_metrics,omitempty"`
		} `json:"runs" binding:"required,dive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if req.DeviceID != deviceInfo.DeviceID {
		utils.ErrorResponse(c, http.StatusForbidden, "Device ID mismatch", "Authenticated device does not match request device_id")
		return
	}

	var results []gin.H
	var successCount, skipCount, errorCount int

	for _, runReq := range req.Runs {
		var existingRun models.Run
		err := h.db.Where("session_id = ?", runReq.SessionID).First(&existingRun).Error
		if err == nil {
			results = append(results, gin.H{
				"session_id": runReq.SessionID,
				"run_id":     existingRun.ID,
				"status":     "already_exists",
			})
			skipCount++
			continue
		}
		if err != gorm.ErrRecordNotFound {
			results = append(results, gin.H{
				"session_id": runReq.SessionID,
				"status":     "error",
				"error":      "Database error: " + err.Error(),
			})
			errorCount++
			continue
		}

		if err := h.validator.Struct(&runReq.RunData); err != nil {
			results = append(results, gin.H{
				"session_id": runReq.SessionID,
				"status":     "error",
				"error":      "Validation error: " + err.Error(),
			})
			errorCount++
			continue
		}

		run := models.Run{
			UserID:          deviceInfo.UserID,
			DeviceID:        req.DeviceID,
			SessionID:       runReq.SessionID,
			Title:           runReq.RunData.Title,
			Notes:           runReq.RunData.Notes,
			StartedAt:       runReq.RunData.StartedAt,
			EndedAt:         runReq.RunData.EndedAt,
			DurationSeconds: runReq.RunData.DurationSeconds,
			DistanceMeters:  runReq.RunData.DistanceMeters,
			AvgSpeedKmh:     runReq.RunData.AvgSpeedKmh,
			MaxSpeedKmh:     runReq.RunData.MaxSpeedKmh,
			CaloriesBurned:  runReq.RunData.CaloriesBurned,
			StepsCount:      runReq.RunData.StepsCount,
			StartLatitude:   runReq.RunData.StartLatitude,
			StartLongitude:  runReq.RunData.StartLongitude,
			EndLatitude:     runReq.RunData.EndLatitude,
			EndLongitude:    runReq.RunData.EndLongitude,
		}

		tx := h.db.Begin()
		if tx.Error != nil {
			results = append(results, gin.H{
				"session_id": runReq.SessionID,
				"status":     "error",
				"error":      "Transaction error: " + tx.Error.Error(),
			})
			errorCount++
			continue
		}

		if err := tx.Create(&run).Error; err != nil {
			tx.Rollback()
			results = append(results, gin.H{
				"session_id": runReq.SessionID,
				"status":     "error",
				"error":      "Failed to create run: " + err.Error(),
			})
			errorCount++
			continue
		}

		if runReq.AIMetrics != nil {
			if err := h.validator.Struct(runReq.AIMetrics); err != nil {
				tx.Rollback()
				results = append(results, gin.H{
					"session_id": runReq.SessionID,
					"status":     "error",
					"error":      "AI metrics validation error: " + err.Error(),
				})
				errorCount++
				continue
			}

			aiMetrics := runReq.AIMetrics.ToModel(run.ID)
			if err := tx.Create(aiMetrics).Error; err != nil {
				tx.Rollback()
				results = append(results, gin.H{
					"session_id": runReq.SessionID,
					"status":     "error",
					"error":      "Failed to save AI metrics: " + err.Error(),
				})
				errorCount++
				continue
			}
		}

		if err := tx.Commit().Error; err != nil {
			results = append(results, gin.H{
				"session_id": runReq.SessionID,
				"status":     "error",
				"error":      "Failed to commit: " + err.Error(),
			})
			errorCount++
			continue
		}

		results = append(results, gin.H{
			"session_id": runReq.SessionID,
			"run_id":     run.ID,
			"status":     "saved",
		})
		successCount++
	}

	now := time.Now()
	deviceInfo.LastSyncAt = &now
	h.db.Save(deviceInfo)

	utils.SuccessResponse(c, http.StatusOK, "Batch upload completed", gin.H{
		"results":       results,
		"summary": gin.H{
			"total":   len(req.Runs),
			"success": successCount,
			"skipped": skipCount,
			"errors":  errorCount,
		},
	})
}

func (h *IoTHandler) UpdateDeviceStatus(c *gin.Context) {
	device, exists := c.Get("device")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Device not found in context", "")
		return
	}

	deviceInfo, ok := device.(*models.Device)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid device context", "")
		return
	}

	var req models.DeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	if req.DeviceID != deviceInfo.DeviceID {
		utils.ErrorResponse(c, http.StatusForbidden, "Device ID mismatch", "Authenticated device does not match request device_id")
		return
	}

	if req.BatteryLevel != nil {
		deviceInfo.BatteryLevel = req.BatteryLevel
	}
	if req.FirmwareVersion != "" {
		deviceInfo.FirmwareVersion = req.FirmwareVersion
	}

	now := time.Now()
	deviceInfo.LastSyncAt = &now

	if err := h.db.Save(deviceInfo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update device status", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Device status updated successfully", gin.H{
		"device_id":   deviceInfo.DeviceID,
		"updated_at":  now.Format(time.RFC3339),
		"status":      "updated",
	})
}

func (h *IoTHandler) GetDeviceConfig(c *gin.Context) {
	device, exists := c.Get("device")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Device not found in context", "")
		return
	}

	deviceInfo, ok := device.(*models.Device)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid device context", "")
		return
	}

	config := models.DeviceConfig{
		UploadIntervalSeconds: 300,
		BatchSize:            10,   
		CompressionEnabled:   true,
	}

	utils.SuccessResponse(c, http.StatusOK, "Device configuration retrieved successfully", gin.H{
		"device_id": deviceInfo.DeviceID,
		"config":    config,
	})
}