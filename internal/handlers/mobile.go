package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/labmino/runsight-backend/internal/models"
	"github.com/labmino/runsight-backend/internal/services"
	"github.com/labmino/runsight-backend/internal/utils"
)

type MobileHandler struct {
	db             *gorm.DB
	validator      *validator.Validate
	pairingService *services.PairingService
}

func NewMobileHandler(db *gorm.DB) *MobileHandler {
	return &MobileHandler{
		db:             db,
		validator:      validator.New(),
		pairingService: services.NewPairingService(db),
	}
}

func (h *MobileHandler) DB() *gorm.DB {
	return h.db
}

func (h *MobileHandler) RequestPairingCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	response, err := h.pairingService.CreatePairingSession(uid)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create pairing session", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pairing code generated successfully", response)
}

func (h *MobileHandler) CheckPairingStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Session ID required", "session_id parameter is missing")
		return
	}

	status, err := h.pairingService.GetPairingStatus(sessionID, uid)
	if err != nil {
		if err.Error() == "pairing session not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Pairing session not found", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get pairing status", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pairing status retrieved successfully", status)
}

func (h *MobileHandler) GetDevices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	var devices []models.Device
	err := h.db.Where("user_id = ? AND is_active = ?", uid, true).Find(&devices).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch devices", err.Error())
		return
	}

	type DeviceResponse struct {
		DeviceID        string `json:"device_id"`
		DeviceName      string `json:"device_name"`
		DeviceType      string `json:"device_type"`
		FirmwareVersion string `json:"firmware_version"`
		IsActive        bool   `json:"is_active"`
		BatteryLevel    *int   `json:"battery_level,omitempty"`
		LastSyncAt      string `json:"last_sync_at,omitempty"`
		PairedAt        string `json:"paired_at"`
	}

	var response []DeviceResponse
	for _, device := range devices {
		deviceResp := DeviceResponse{
			DeviceID:        device.DeviceID,
			DeviceName:      device.DeviceName,
			DeviceType:      device.DeviceType,
			FirmwareVersion: device.FirmwareVersion,
			IsActive:        device.IsActive,
			BatteryLevel:    device.BatteryLevel,
			PairedAt:        device.PairedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if device.LastSyncAt != nil {
			deviceResp.LastSyncAt = device.LastSyncAt.Format("2006-01-02T15:04:05Z07:00")
		}
		response = append(response, deviceResp)
	}

	utils.SuccessResponse(c, http.StatusOK, "Devices retrieved successfully", response)
}

func (h *MobileHandler) RemoveDevice(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	deviceID := c.Param("device_id")
	if deviceID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Device ID required", "device_id parameter is missing")
		return
	}

	var device models.Device
	err := h.db.Where("device_id = ? AND user_id = ? AND is_active = ?", deviceID, uid, true).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Device not found", "Device not found or already removed")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	device.IsActive = false
	if err := h.db.Save(&device).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove device", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Device removed successfully", gin.H{
		"device_id": deviceID,
		"status":    "removed",
	})
}

func (h *MobileHandler) ListRuns(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	query := h.db.Where("user_id = ?", uid)

	if startDate := c.Query("start_date"); startDate != "" {
		if parsedStartDate, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("started_at >= ?", parsedStartDate)
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if parsedEndDate, err := time.Parse("2006-01-02", endDate); err == nil {
			endOfDay := parsedEndDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("started_at <= ?", endOfDay)
		}
	}

	var total int64
	if err := query.Model(&models.Run{}).Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count runs", err.Error())
		return
	}

	var runs []models.Run
	err := query.Order("started_at DESC").Limit(limit).Offset(offset).Find(&runs).Error
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch runs", err.Error())
		return
	}

	type RunListResponse struct {
		ID              uuid.UUID  `json:"id"`
		DeviceID        string     `json:"device_id"`
		SessionID       string     `json:"session_id"`
		Title           string     `json:"title,omitempty"`
		StartedAt       time.Time  `json:"started_at"`
		EndedAt         *time.Time `json:"ended_at,omitempty"`
		DurationSeconds *int       `json:"duration_seconds,omitempty"`
		DistanceMeters  *float64   `json:"distance_meters,omitempty"`
		AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty"`
		CaloriesBurned  *int       `json:"calories_burned,omitempty"`
		CreatedAt       time.Time  `json:"created_at"`
	}

	var response []RunListResponse
	for _, run := range runs {
		response = append(response, RunListResponse{
			ID:              run.ID,
			DeviceID:        run.DeviceID,
			SessionID:       run.SessionID,
			Title:           run.Title,
			StartedAt:       run.StartedAt,
			EndedAt:         run.EndedAt,
			DurationSeconds: run.DurationSeconds,
			DistanceMeters:  run.DistanceMeters,
			AvgSpeedKmh:     run.AvgSpeedKmh,
			CaloriesBurned:  run.CaloriesBurned,
			CreatedAt:       run.CreatedAt,
		})
	}

	totalPages := (int(total) + limit - 1) / limit

	utils.SuccessResponse(c, http.StatusOK, "Runs retrieved successfully", gin.H{
		"runs": response,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  total,
			"limit":        limit,
		},
	})
}

func (h *MobileHandler) GetRun(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	runIDParam := c.Param("run_id")
	if runIDParam == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Run ID required", "run_id parameter is missing")
		return
	}

	runID, err := uuid.Parse(runIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid run ID", "run_id must be a valid UUID")
		return
	}

	var run models.Run
	err = h.db.Where("id = ? AND user_id = ?", runID, uid).First(&run).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Run not found", "Run not found or access denied")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	var aiMetrics models.AIMetrics
	hasAIMetrics := false
	err = h.db.Where("run_id = ?", runID).First(&aiMetrics).Error
	if err == nil {
		hasAIMetrics = true
	} else if err != gorm.ErrRecordNotFound {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch AI metrics", err.Error())
		return
	}

	type RunDetailResponse struct {
		models.Run
		AIMetrics *models.AIMetrics `json:"ai_metrics,omitempty"`
	}

	response := RunDetailResponse{
		Run: run,
	}

	if hasAIMetrics {
		response.AIMetrics = &aiMetrics
	}

	utils.SuccessResponse(c, http.StatusOK, "Run retrieved successfully", response)
}

func (h *MobileHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	type Stats struct {
		TotalDistance   float64 `json:"total_distance_meters"`
		TotalRuns       int64   `json:"total_runs"`
		TotalDuration   int     `json:"total_duration_seconds"`
		AvgSpeed        float64 `json:"avg_speed_kmh"`
		TotalCalories   int     `json:"total_calories_burned"`
		TotalSteps      int     `json:"total_steps"`
	}

	var stats Stats

	err := h.db.Model(&models.Run{}).
		Select(`
			COUNT(*) as total_runs,
			COALESCE(SUM(distance_meters), 0) as total_distance,
			COALESCE(SUM(duration_seconds), 0) as total_duration,
			COALESCE(AVG(avg_speed_kmh), 0) as avg_speed,
			COALESCE(SUM(calories_burned), 0) as total_calories,
			COALESCE(SUM(steps_count), 0) as total_steps
		`).
		Where("user_id = ?", uid).
		Scan(&stats).Error

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch statistics", err.Error())
		return
	}

	type StatsResponse struct {
		TotalDistanceMeters   float64 `json:"total_distance_meters"`
		TotalDistanceKm       float64 `json:"total_distance_km"`
		TotalRuns             int64   `json:"total_runs"`
		TotalDurationSeconds  int     `json:"total_duration_seconds"`
		TotalDurationHours    float64 `json:"total_duration_hours"`
		AvgSpeedKmh          float64 `json:"avg_speed_kmh"`
		TotalCaloriesBurned  int     `json:"total_calories_burned"`
		TotalSteps           int     `json:"total_steps"`
		AvgDistancePerRun    float64 `json:"avg_distance_per_run_meters"`
		AvgDurationPerRun    float64 `json:"avg_duration_per_run_seconds"`
	}

	var avgDistance, avgDuration float64
	if stats.TotalRuns > 0 {
		avgDistance = stats.TotalDistance / float64(stats.TotalRuns)
		avgDuration = float64(stats.TotalDuration) / float64(stats.TotalRuns)
	}

	response := StatsResponse{
		TotalDistanceMeters:   stats.TotalDistance,
		TotalDistanceKm:       stats.TotalDistance / 1000,
		TotalRuns:             stats.TotalRuns,
		TotalDurationSeconds:  stats.TotalDuration,
		TotalDurationHours:    float64(stats.TotalDuration) / 3600,
		AvgSpeedKmh:          stats.AvgSpeed,
		TotalCaloriesBurned:  stats.TotalCalories,
		TotalSteps:           stats.TotalSteps,
		AvgDistancePerRun:    avgDistance,
		AvgDurationPerRun:    avgDuration,
	}

	utils.SuccessResponse(c, http.StatusOK, "Statistics retrieved successfully", response)
}

func (h *MobileHandler) UpdateRunNotes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User ID not found")
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", "User ID type assertion failed")
		return
	}

	runIDParam := c.Param("run_id")
	if runIDParam == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Run ID required", "run_id parameter is missing")
		return
	}

	runID, err := uuid.Parse(runIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid run ID", "run_id must be a valid UUID")
		return
	}

	type UpdateRunNotesRequest struct {
		Title *string `json:"title,omitempty" validate:"omitempty,max=100"`
		Notes *string `json:"notes,omitempty"`
	}

	var req UpdateRunNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var run models.Run
	err = h.db.Where("id = ? AND user_id = ?", runID, uid).First(&run).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Run not found", "Run not found or access denied")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	updated := false

	if req.Title != nil {
		run.Title = *req.Title
		updated = true
	}

	if req.Notes != nil {
		run.Notes = *req.Notes
		updated = true
	}

	if !updated {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update", "At least one field (title or notes) must be provided")
		return
	}

	if err := h.db.Save(&run).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update run", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Run updated successfully", gin.H{
		"run_id":     run.ID,
		"title":      run.Title,
		"notes":      run.Notes,
		"updated_at": run.UpdatedAt,
	})
}