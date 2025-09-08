package handlers

import (
	"net/http"

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