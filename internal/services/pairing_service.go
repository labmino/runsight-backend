package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/labmino/runsight-backend/internal/models"
	"github.com/labmino/runsight-backend/internal/utils"
)

type PairingService struct {
	db *gorm.DB
}

func NewPairingService(db *gorm.DB) *PairingService {
	return &PairingService{db: db}
}

func (s *PairingService) CreatePairingSession(userID uuid.UUID) (*models.PairingResponse, error) {
	code := utils.GeneratePairingCode()
	
	// Ensure pairing code is unique among active sessions
	for {
		var existing models.PairingSession
		err := s.db.Where("code = ? AND status = ? AND expires_at > ?", 
			code, models.PairingStatusPending, time.Now()).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("database error checking code uniqueness: %w", err)
		}
		code = utils.GeneratePairingCode()
	}

	session := models.PairingSession{
		UserID: userID,
		Code:   code,
		Status: models.PairingStatusPending,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create pairing session: %w", err)
	}

	return &models.PairingResponse{
		Code:             session.Code,
		SessionID:        session.ID,
		ExpiresAt:        session.ExpiresAt.Format(time.RFC3339),
		ExpiresInSeconds: session.RemainingSeconds(),
	}, nil
}

func (s *PairingService) GetPairingStatus(sessionID string, userID uuid.UUID) (*models.PairingStatusResponse, error) {
	var session models.PairingSession
	err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("pairing session not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if session.IsExpired() {
		if session.Status != models.PairingStatusExpired {
			session.Status = models.PairingStatusExpired
			s.db.Save(&session)
		}
		
		return &models.PairingStatusResponse{
			Paired:           false,
			Expired:          true,
			RemainingSeconds: 0,
		}, nil
	}

	if session.Status == models.PairingStatusPaired && session.DeviceID != "" {
		var device models.Device
		err := s.db.Where("device_id = ?", session.DeviceID).First(&device).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch device info: %w", err)
		}

		return &models.PairingStatusResponse{
			Paired:  true,
			Expired: false,
			Device: &models.PairingDeviceInfo{
				DeviceID:        device.DeviceID,
				DeviceName:      device.DeviceName,
				DeviceType:      device.DeviceType,
				FirmwareVersion: device.FirmwareVersion,
				PairedAt:        device.PairedAt,
			},
		}, nil
	}

	return &models.PairingStatusResponse{
		Paired:           false,
		Expired:          false,
		RemainingSeconds: session.RemainingSeconds(),
	}, nil
}

func (s *PairingService) VerifyPairingCode(req *models.DeviceRegisterRequest) (*models.Device, error) {
	var session models.PairingSession
	err := s.db.Where("code = ? AND status = ? AND expires_at > ?", 
		req.Code, models.PairingStatusPending, time.Now()).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired pairing code")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	var existingDevice models.Device
	err = s.db.Where("device_id = ?", req.DeviceID).First(&existingDevice).Error
	if err == nil {
		return nil, errors.New("device already registered")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error checking existing device: %w", err)
	}

	deviceToken, err := utils.GenerateDeviceToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate device token: %w", err)
	}

	device := models.Device{
		DeviceID:        req.DeviceID,
		UserID:          session.UserID,
		DeviceName:      req.DeviceID,
		DeviceType:      req.DeviceType,
		FirmwareVersion: req.FirmwareVersion,
		HardwareVersion: req.HardwareVersion,
		MACAddress:      req.MACAddress,
		DeviceToken:     deviceToken,
		IsActive:        true,
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	if err := tx.Create(&device).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	now := time.Now()
	session.DeviceID = req.DeviceID
	session.Status = models.PairingStatusPaired
	session.PairedAt = &now
	
	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update pairing session: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &device, nil
}

func (s *PairingService) CleanupExpiredSessions() error {
	result := s.db.Where("expires_at < ?", time.Now()).
		Delete(&models.PairingSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}
	
	return nil
}